package koyeb

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	// DefaultClaimWaitTimeout and DefaultClaimPollInterval mirror the
	// Python SDK's wait_claim_ready defaults.
	DefaultClaimWaitTimeout  = 300 * time.Second
	DefaultClaimPollInterval = 2 * time.Second
)

// serviceStatusClass classifies a service status for claim readiness.
type serviceStatusClass int

const (
	serviceStatusInProgress serviceStatusClass = iota
	serviceStatusReady
	serviceStatusTerminal
)

// classifyServiceStatus fails closed like the SDKs: HEALTHY/DEGRADED are
// ready, STARTING/RESUMING are in progress, everything else — including
// unknown forward-compat values — is terminal.
func classifyServiceStatus(status koyeb.ServiceStatus) serviceStatusClass {
	switch status {
	case koyeb.SERVICESTATUS_HEALTHY, koyeb.SERVICESTATUS_DEGRADED:
		return serviceStatusReady
	case koyeb.SERVICESTATUS_STARTING, koyeb.SERVICESTATUS_RESUMING:
		return serviceStatusInProgress
	default:
		return serviceStatusTerminal
	}
}

// waitClaimReady polls the claimed service until it is ready, mirroring
// the SDKs' wait_claim_ready: transient GetService failures are treated as
// in progress and retried until the timeout, terminal states error out.
func waitClaimReady(ctx context.Context, serviceID string, timeout, pollInterval time.Duration,
	getStatus func(context.Context, string) (koyeb.ServiceStatus, error),
) error {
	deadline := time.Now().Add(timeout)

	for {
		status, err := getStatus(ctx, serviceID)
		if err == nil {
			switch classifyServiceStatus(status) {
			case serviceStatusReady:
				return nil
			case serviceStatusTerminal:
				return &errors.CLIError{
					What:     "Claimed service reached a terminal state",
					Why:      fmt.Sprintf("Service '%s' reached terminal state '%s' and will not become ready.", serviceID, status),
					Solution: errors.CLIErrorSolution("Check the service logs with `koyeb service logs " + serviceID + "`"),
				}
			}
		}

		if !time.Now().Before(deadline) {
			return &errors.CLIError{
				What:     "Timed out waiting for the claimed service",
				Why:      fmt.Sprintf("service %s did not become ready within %s", serviceID, timeout),
				Solution: errors.CLIErrorSolution("Check the service status with `koyeb service get " + serviceID + "`"),
			}
		}

		select {
		case <-time.After(pollInterval):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// generateRequestID returns a UUID v7 used when --request-id is not given.
func generateRequestID() string {
	return uuid.NewV7().String()
}

// buildPoolClaimRequest builds the PoolClaimRequest sent to the claim API.
func buildPoolClaimRequest(poolID, requestID string) koyeb.PoolClaimRequest {
	return koyeb.PoolClaimRequest{
		PoolId:    &poolID,
		RequestId: &requestID,
	}
}

func newPoolClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim POOL",
		Short: "Claim an instance from a pool",
		Long: `Claim an instance from a pool.

When --request-id is not provided, a random UUID v7 is generated. Replaying
a claim with the same request ID is idempotent: the server returns the
previously created claim instead of provisioning a new instance.`,
		Args: cobra.ExactArgs(1),
		Example: `
# Claim an instance from a pool
$> koyeb pool claim my-pool

# Claim with an explicit request ID (idempotent retries)
$> koyeb pool claim my-pool --request-id my-request-id
`,
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			if err := setProjectHeader(ctx, cmd); err != nil {
				return err
			}
			poolID, err := ResolvePoolArgs(ctx, args[0])
			if err != nil {
				return err
			}

			requestID, _ := cmd.Flags().GetString("request-id")
			if requestID == "" {
				requestID = generateRequestID()
			}

			req := buildPoolClaimRequest(poolID, requestID)

			res, resp, err := ctx.Client.PoolClaimsApi.Claim(ctx.Context).Body(req).Execute()
			if err != nil {
				return errors.NewCLIErrorFromAPIError(
					fmt.Sprintf("Error while claiming an instance from the pool `%s`", args[0]),
					err,
					resp,
				)
			}

			full := GetBoolFlags(cmd, "full")
			claimReply := NewClaimReply(ctx.Mapper, res, full)
			ctx.Renderer.Render(claimReply)

			if GetBoolFlags(cmd, "wait") {
				if serviceID := res.GetServiceId(); serviceID != "" {
					if err := waitClaimedService(ctx, serviceID); err != nil {
						return err
					}
					log.Infof("Claimed service %s is ready", serviceID)
				}
			}
			return nil
		}),
	}
	cmd.Flags().String("request-id", "", "Claim request ID (defaults to a generated UUID v4)")
	cmd.Flags().Bool("wait", false, "Wait until the claimed service is ready (timeout 5m, poll 2s)")

	return cmd
}

// waitClaimedService polls GetService until the claimed service is ready.
// Transient API failures count as in progress, matching the SDKs.
func waitClaimedService(ctx *CLIContext, serviceID string) error {
	getStatus := func(c context.Context, id string) (koyeb.ServiceStatus, error) {
		res, _, err := ctx.Client.ServicesApi.GetService(c, id).Execute()
		if err != nil {
			return "", err
		}
		service := res.GetService()
		if !service.HasStatus() {
			return "", fmt.Errorf("service %s has no status", id)
		}
		return service.GetStatus(), nil
	}

	return waitClaimReady(ctx.Context, serviceID, DefaultClaimWaitTimeout, DefaultClaimPollInterval, getStatus)
}
