package koyeb

import (
	"fmt"
	"testing"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sandboxSnapshotCreateCmd(t *testing.T) *cobra.Command {
	t.Helper()
	cmd, _, err := NewSandboxCmd().Find([]string{"snapshot", "create"})
	require.NoError(t, err)
	return cmd
}

func TestSandboxSnapshotCreateSubcommand(t *testing.T) {
	cmd := sandboxSnapshotCreateCmd(t)
	flags := cmd.Flags()

	assert.Equal(t, "create NAME", cmd.Use)
	assert.NoError(t, cmd.Args(nil, []string{"sb"}), "snapshot create takes exactly one sandbox name")
	assert.Error(t, cmd.Args(nil, nil), "snapshot create requires a sandbox name")

	assert.Equal(t, "", flags.Lookup("name").DefValue, "the snapshot name defaults to a generated one")
	assert.Equal(t, "filesystem", flags.Lookup("type").DefValue)

	// --wait stays opt-in (intentional CLI UX) but must note the SDK default.
	waitFlag := flags.Lookup("wait")
	require.NotNil(t, waitFlag)
	assert.Equal(t, "false", waitFlag.DefValue, "--wait must stay opt-in")
	assert.Contains(t, waitFlag.Usage, "SDKs wait by default")

	waitTimeout, err := time.ParseDuration(flags.Lookup("wait-timeout").DefValue)
	require.NoError(t, err)
	assert.Equal(t, 10*time.Minute, waitTimeout, "the SDKs wait up to 600s")
	assert.Equal(t, "5", flags.Lookup("poll-interval").DefValue, "the SDKs poll every 5s")
}

// snapshotFixtureAPI wires the sandbox fixture with one running instance and
// a snapshot the create call can return.
func snapshotFixtureAPI() *fakeAPI {
	fake := sandboxFixtureAPI()
	instanceID := "823e4567-e89b-42d3-a456-426614174000"
	creating := koyeb.INSTANCESNAPSHOTSTATUS_CREATING
	filesystem := koyeb.INSTANCESNAPSHOTTYPE_FILESYSTEM
	fake.instances = []koyeb.InstanceListItem{{Id: &instanceID}}
	fake.createSnapshotReply = &koyeb.InstanceSnapshot{
		Id:     koyeb.PtrString("923e4567-e89b-42d3-a456-426614174000"),
		Name:   koyeb.PtrString("my-snap"),
		Status: &creating,
		Type:   &filesystem,
	}
	return fake
}

func runSnapshotCreate(t *testing.T, fake *fakeAPI, cmd *cobra.Command) (string, error) {
	t.Helper()
	ctx := sandboxTestContext(fake)

	var handlerErr error
	out, _ := captureStd(t, func() {
		handlerErr = NewSandboxHandler().SnapshotCreate(ctx, cmd, []string{"323e4567-e89b-42d3-a456-426614174000"})
	})
	return out, handlerErr
}

func TestSandboxSnapshotCreateFlow(t *testing.T) {
	t.Run("snapshots the first running instance", func(t *testing.T) {
		fake := snapshotFixtureAPI()
		cmd := sandboxSnapshotCreateCmd(t)
		require.NoError(t, cmd.Flags().Set("name", "my-snap"))

		out, err := runSnapshotCreate(t, fake, cmd)

		require.NoError(t, err)
		require.NotNil(t, fake.createSnapshotReq)
		assert.Equal(t, "823e4567-e89b-42d3-a456-426614174000", fake.createSnapshotReq.GetInstanceId(),
			"the snapshot targets the sandbox's running instance")
		assert.Equal(t, "my-snap", fake.createSnapshotReq.GetName())
		assert.Equal(t, koyeb.INSTANCESNAPSHOTTYPE_FILESYSTEM, fake.createSnapshotReq.GetType())
		assert.Equal(t, "323e4567-e89b-42d3-a456-426614174000", fake.listInstancesServiceID)
		assert.Equal(t, []string{"ALLOCATING", "STARTING", "HEALTHY"}, fake.listInstancesStatuses,
			"the same instance statuses the SDK snapshots from")
		assert.Equal(t, "1", fake.listInstancesLimit)
		assert.Contains(t, out, "923e4567-e89b-42d3-a456-426614174000")
		assert.Contains(t, out, "my-snap")
		assert.Contains(t, out, "CREATING", "the status is reported without --wait")
	})

	t.Run("generates a snapshot name when --name is empty", func(t *testing.T) {
		fake := snapshotFixtureAPI()
		cmd := sandboxSnapshotCreateCmd(t)

		_, err := runSnapshotCreate(t, fake, cmd)

		require.NoError(t, err)
		require.NotNil(t, fake.createSnapshotReq)
		name := fake.createSnapshotReq.GetName()
		assert.Contains(t, name, "323e4567", "the generated name derives from the sandbox reference")
		assert.Contains(t, name, fmt.Sprintf("snapshot-%d", time.Now().Unix()), "the generated name is unique per second")
	})

	t.Run("--type full requests a FULL snapshot", func(t *testing.T) {
		fake := snapshotFixtureAPI()
		cmd := sandboxSnapshotCreateCmd(t)
		require.NoError(t, cmd.Flags().Set("type", "full"))

		_, err := runSnapshotCreate(t, fake, cmd)

		require.NoError(t, err)
		assert.Equal(t, koyeb.INSTANCESNAPSHOTTYPE_FULL, fake.createSnapshotReq.GetType())
	})

	t.Run("rejects an invalid --type before any API call", func(t *testing.T) {
		fake := snapshotFixtureAPI()
		cmd := sandboxSnapshotCreateCmd(t)
		require.NoError(t, cmd.Flags().Set("type", "bogus"))

		_, err := runSnapshotCreate(t, fake, cmd)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--type")
		assert.Empty(t, fake.servicesFetched, "flag validation must precede the sandbox resolution")
		assert.Nil(t, fake.createSnapshotReq, "no snapshot may be created")
	})

	t.Run("rejects a non-sandbox service", func(t *testing.T) {
		fake := snapshotFixtureAPI()
		webType := koyeb.SERVICETYPE_WEB
		fake.service.Type = &webType
		cmd := sandboxSnapshotCreateCmd(t)

		_, err := runSnapshotCreate(t, fake, cmd)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a sandbox")
		assert.Nil(t, fake.createSnapshotReq)
	})

	t.Run("fails when no instance is running", func(t *testing.T) {
		fake := snapshotFixtureAPI()
		fake.instances = nil
		cmd := sandboxSnapshotCreateCmd(t)

		_, err := runSnapshotCreate(t, fake, cmd)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no running instance")
		assert.Nil(t, fake.createSnapshotReq)
	})

	t.Run("create API errors surface", func(t *testing.T) {
		fake := snapshotFixtureAPI()
		fake.createSnapshotErr = fmt.Errorf("quota exceeded")
		cmd := sandboxSnapshotCreateCmd(t)

		_, err := runSnapshotCreate(t, fake, cmd)

		require.Error(t, err)
	})
}

func TestDefaultSnapshotName(t *testing.T) {
	t.Run("derives from the service part of app/name", func(t *testing.T) {
		name := defaultSnapshotName("myapp/mysandbox")
		assert.Contains(t, name, "mysandbox-snapshot-", "the service part names the snapshot")
		assert.NotContains(t, name, "myapp", "the app part must not leak into the name")
	})

	t.Run("uses the whole reference without a slash", func(t *testing.T) {
		assert.Contains(t, defaultSnapshotName("323e4567-e89b-42d3-a456-426614174000"),
			"323e4567-e89b-42d3-a456-426614174000-snapshot-")
	})
}

func TestSandboxSnapshotCreateWaitFlow(t *testing.T) {
	// waitCmd returns a snapshot create command set up for fast polls.
	waitCmd := func(t *testing.T) *cobra.Command {
		t.Helper()
		cmd := sandboxSnapshotCreateCmd(t)
		require.NoError(t, cmd.Flags().Set("wait", "true"))
		require.NoError(t, cmd.Flags().Set("poll-interval", "0.01"))
		return cmd
	}

	t.Run("waits until the snapshot is available", func(t *testing.T) {
		fake := snapshotFixtureAPI()
		fake.snapshotStatuses = []koyeb.InstanceSnapshotStatus{
			koyeb.INSTANCESNAPSHOTSTATUS_CREATING,
			koyeb.INSTANCESNAPSHOTSTATUS_CREATING,
			koyeb.INSTANCESNAPSHOTSTATUS_AVAILABLE,
		}
		cmd := waitCmd(t)
		require.NoError(t, cmd.Flags().Set("wait-timeout", "5s"), "bound the wait so a regression fails fast")

		out, err := runSnapshotCreate(t, fake, cmd)

		require.NoError(t, err)
		assert.Contains(t, out, "Status: "+string(koyeb.INSTANCESNAPSHOTSTATUS_AVAILABLE))
		assert.GreaterOrEqual(t, fake.snapshotCalls, 3, "the wait must poll until available")
	})

	t.Run("transient fetch errors keep polling", func(t *testing.T) {
		// A failed status fetch says nothing about the snapshot: the wait
		// must keep polling, not succeed on the error.
		fake := snapshotFixtureAPI()
		fake.snapshotGetErrs = []error{fmt.Errorf("transient fetch failure"), nil}
		fake.snapshotStatuses = []koyeb.InstanceSnapshotStatus{koyeb.INSTANCESNAPSHOTSTATUS_AVAILABLE}
		cmd := waitCmd(t)
		require.NoError(t, cmd.Flags().Set("wait-timeout", "5s"), "bound the wait so a regression fails fast")

		out, err := runSnapshotCreate(t, fake, cmd)

		require.NoError(t, err)
		assert.Contains(t, out, "Status: "+string(koyeb.INSTANCESNAPSHOTSTATUS_AVAILABLE))
		assert.GreaterOrEqual(t, fake.snapshotCalls, 2, "a transient fetch error must not end the wait")
	})

	t.Run("ERROR is terminal", func(t *testing.T) {
		fake := snapshotFixtureAPI()
		fake.snapshotStatuses = []koyeb.InstanceSnapshotStatus{koyeb.INSTANCESNAPSHOTSTATUS_ERROR}
		cmd := waitCmd(t)
		require.NoError(t, cmd.Flags().Set("wait-timeout", "2s"), "bound the wait so a regression fails fast")

		_, err := runSnapshotCreate(t, fake, cmd)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "terminal state")
	})

	t.Run("times out when the snapshot never becomes available", func(t *testing.T) {
		fake := snapshotFixtureAPI()
		fake.snapshotStatuses = []koyeb.InstanceSnapshotStatus{koyeb.INSTANCESNAPSHOTSTATUS_CREATING}
		cmd := waitCmd(t)
		require.NoError(t, cmd.Flags().Set("wait-timeout", "40ms"))

		start := time.Now()
		_, err := runSnapshotCreate(t, fake, cmd)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "did not become available")
		assert.Less(t, time.Since(start), 5*time.Second, "the timeout must bound the wait")
	})

	t.Run("an invalid --wait-timeout fails before any API call", func(t *testing.T) {
		fake := snapshotFixtureAPI()
		cmd := sandboxSnapshotCreateCmd(t)
		require.NoError(t, cmd.Flags().Set("wait", "true"))
		require.NoError(t, cmd.Flags().Set("wait-timeout", "0"))

		_, err := runSnapshotCreate(t, fake, cmd)

		require.Error(t, err)
		assert.Empty(t, fake.servicesFetched, "flag validation must precede the sandbox resolution")
		assert.Nil(t, fake.createSnapshotReq)
	})
}

func TestSnapshotBootsSandbox(t *testing.T) {
	t.Run("the created snapshot resolves by ID and boots via --snapshot", func(t *testing.T) {
		fake := snapshotFixtureAPI()
		available := koyeb.INSTANCESNAPSHOTSTATUS_AVAILABLE
		fake.createSnapshotReply.Status = &available
		ctx := sandboxTestContext(fake)

		fake.snapshot = fake.createSnapshotReply
		snapshotID, snapshotType := resolveSnapshotFlags(ctx, "923e4567-e89b-42d3-a456-426614174000")
		assert.Equal(t, "923e4567-e89b-42d3-a456-426614174000", snapshotID)
		assert.Equal(t, koyeb.INSTANCESNAPSHOTTYPE_FILESYSTEM, snapshotType)

		createService := koyeb.NewCreateServiceWithDefaults()
		wireSnapshot(createService, snapshotID, snapshotType, "sb")
		assert.Equal(t, "923e4567-e89b-42d3-a456-426614174000", createService.GetInstanceSnapshotId())
	})

	t.Run("the created snapshot resolves by name", func(t *testing.T) {
		fake := snapshotFixtureAPI()
		available := koyeb.INSTANCESNAPSHOTSTATUS_AVAILABLE
		fake.createSnapshotReply.Status = &available
		ctx := sandboxTestContext(fake)

		fake.snapshots = []koyeb.InstanceSnapshot{*fake.createSnapshotReply}
		snapshotID, snapshotType := resolveSnapshotFlags(ctx, "my-snap")
		assert.Equal(t, "923e4567-e89b-42d3-a456-426614174000", snapshotID)
		assert.Equal(t, koyeb.INSTANCESNAPSHOTTYPE_FILESYSTEM, snapshotType)
	})
}
