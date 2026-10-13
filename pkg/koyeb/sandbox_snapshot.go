package koyeb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

// snapshotableInstanceStatuses lists the instance statuses a snapshot can
// be taken from — the same set the Python SDK's snapshot() filters on.
var snapshotableInstanceStatuses = []string{
	string(koyeb.INSTANCESTATUS_ALLOCATING),
	string(koyeb.INSTANCESTATUS_STARTING),
	string(koyeb.INSTANCESTATUS_HEALTHY),
}

// SnapshotCreate captures an instance snapshot of a running sandbox — the
// boot source 'koyeb sandbox create --snapshot' and the SDKs' snapshot()
// and spawn() consume.
func (h *SandboxHandler) SnapshotCreate(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	sandboxName := args[0]

	// Validate the flags before any API call, so a typo creates nothing.
	snapshotType, err := parseSnapshotType(GetStringFlags(cmd, "type"))
	if err != nil {
		return err
	}
	if GetBoolFlags(cmd, "wait") {
		if _, err := waitTimeoutFlag(cmd); err != nil {
			return err
		}
	}

	service, err := resolveSandboxService(ctx, sandboxName)
	if err != nil {
		return err
	}
	serviceID := service.GetId()

	instanceID, err := runningInstanceID(ctx, serviceID)
	if err != nil {
		return err
	}

	name := GetStringFlags(cmd, "name")
	if name == "" {
		name = defaultSnapshotName(sandboxName)
	}

	req := koyeb.NewCreateInstanceSnapshotRequestWithDefaults()
	req.SetInstanceId(instanceID)
	req.SetName(name)
	req.SetType(snapshotType)

	reply, resp, err := ctx.API.CreateInstanceSnapshot(ctx.Context, *req)
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while creating an instance snapshot of sandbox `%s`", sandboxName),
			err,
			resp,
		)
	}
	snapshot := reply.GetInstanceSnapshot()

	if GetBoolFlags(cmd, "wait") {
		if err := waitForSnapshotAvailable(ctx, cmd, snapshot.GetId()); err != nil {
			return err
		}
	}

	printSnapshotCreateResult(cmd, snapshot, name, GetStringFlags(cmd, "type"))
	return nil
}

// printSnapshotCreateResult reports the created snapshot like the rest of
// the sandbox subtree: plain text, one field per line. After a successful
// --wait the status is AVAILABLE — the wait's contract, no re-fetch needed.
func printSnapshotCreateResult(cmd *cobra.Command, snapshot koyeb.InstanceSnapshot, name, typeFlag string) {
	fmt.Printf("Snapshot ID: %s\n", snapshot.GetId())
	fmt.Printf("Name: %s\n", name)
	fmt.Printf("Type: %s\n", typeFlag)
	if GetBoolFlags(cmd, "wait") {
		fmt.Printf("Status: %s\n", koyeb.INSTANCESNAPSHOTSTATUS_AVAILABLE)
		return
	}
	fmt.Printf("Status: %s\n", snapshot.GetStatus())
}

// parseSnapshotType maps the --type flag to the API enum; the empty string
// keeps the SDK default (FILESYSTEM).
func parseSnapshotType(value string) (koyeb.InstanceSnapshotType, error) {
	switch value {
	case "", "filesystem":
		return koyeb.INSTANCESNAPSHOTTYPE_FILESYSTEM, nil
	case "full":
		return koyeb.INSTANCESNAPSHOTTYPE_FULL, nil
	}
	return "", &errors.CLIError{
		What:     "Invalid --type",
		Why:      fmt.Sprintf("'%s' is not a valid snapshot type", value),
		Solution: "Use 'filesystem' or 'full'",
	}
}

// defaultSnapshotName derives a snapshot name from the sandbox reference —
// the service part of "app/name" plus a second-precision timestamp, so
// successive snapshots of one sandbox do not collide.
func defaultSnapshotName(sandboxRef string) string {
	_, service, found := strings.Cut(sandboxRef, "/")
	if !found {
		service = sandboxRef
	}
	return fmt.Sprintf("%s-snapshot-%d", service, time.Now().Unix())
}

// runningInstanceID returns the ID of the instance serving the sandbox —
// the first instance in a snapshotable state, like the Python SDK's
// snapshot().
func runningInstanceID(ctx *CLIContext, serviceID string) (string, error) {
	reply, resp, err := ctx.API.ListInstancesByService(ctx.Context, serviceID, snapshotableInstanceStatuses, "1")
	if err != nil {
		return "", errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while listing the instances of service %s", serviceID),
			err,
			resp,
		)
	}

	instances := reply.GetInstances()
	if len(instances) == 0 || instances[0].GetId() == "" {
		return "", &errors.CLIError{
			What:     "No running instance to snapshot",
			Why:      fmt.Sprintf("service %s has no running instance", serviceID),
			Solution: "Wait for the sandbox to be deployed and try again",
		}
	}
	return instances[0].GetId(), nil
}

// waitForSnapshotAvailable polls the snapshot until AVAILABLE, mirroring
// the SDKs' wait_available: ERROR fails fast, a timeout fails the command.
func waitForSnapshotAvailable(ctx *CLIContext, cmd *cobra.Command, snapshotID string) error {
	waitTimeout, err := waitTimeoutFlag(cmd)
	if err != nil {
		return err
	}

	terminalErr := func(status koyeb.InstanceSnapshotStatus) error {
		return &errors.CLIError{
			What: "Snapshot failed",
			Why: fmt.Sprintf("instance snapshot %s reached terminal state %s and will not become available",
				snapshotID, status),
			Solution: "Check the sandbox, then create the snapshot again",
		}
	}
	timeoutErr := func() error {
		return &errors.CLIError{
			What:     "Timed out waiting for the instance snapshot",
			Why:      fmt.Sprintf("instance snapshot %s did not become available within %s", snapshotID, waitTimeout),
			Solution: errors.CLIErrorSolution("Raise --wait-timeout, or check the snapshot status again later"),
		}
	}

	return waitEngine(ctx.Context, waitTimeout, waitPollInterval(cmd),
		snapshotAvailableProbe(ctx, snapshotID, terminalErr),
		timeoutErr,
	)
}

// snapshotAvailableProbe polls the snapshot status with the SDKs'
// classification: AVAILABLE is done, ERROR is terminal, every other status
// keeps polling. Transient fetch errors retry until the timeout, like the
// sandbox deployment wait.
func snapshotAvailableProbe(ctx *CLIContext, snapshotID string, terminalErr func(koyeb.InstanceSnapshotStatus) error) waitProbe {
	return func(c context.Context) (bool, error) {
		reply, _, err := ctx.API.GetInstanceSnapshot(c, snapshotID)
		if err != nil {
			return false, nil
		}
		snapshot := reply.GetInstanceSnapshot()
		switch snapshot.GetStatus() {
		case koyeb.INSTANCESNAPSHOTSTATUS_AVAILABLE:
			return true, nil
		case koyeb.INSTANCESNAPSHOTSTATUS_ERROR:
			return false, terminalErr(koyeb.INSTANCESNAPSHOTSTATUS_ERROR)
		}
		return false, nil
	}
}
