package koyeb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sandboxCreateCmd(t *testing.T) *cobra.Command {
	t.Helper()
	cmd, _, err := NewSandboxCmd().Find([]string{"create"})
	require.NoError(t, err)
	return cmd
}

func parseSandboxDefinition(t *testing.T, args []string) (*koyeb.DeploymentDefinition, error) {
	t.Helper()
	cmd := sandboxCreateCmd(t)
	require.NoError(t, cmd.Flags().Parse(args))
	def := koyeb.NewDeploymentDefinitionWithDefaults()
	err := parseSandboxDefinitionFlags(&CLIContext{}, cmd, def, NewServiceHandler())
	return def, err
}

func TestSandboxCreateFlagsRegistered(t *testing.T) {
	flags := sandboxCreateCmd(t).Flags()

	for _, name := range []string{
		"enable-mesh", "exposed-port-protocol", "enable-tcp-proxy",
		"sandbox-secret", "poll-interval", "cleanup-on-failure", "snapshot",
	} {
		assert.NotNil(t, flags.Lookup(name), "flag --%s must be registered on sandbox create", name)
	}

	protocol := flags.Lookup("exposed-port-protocol")
	require.NotNil(t, protocol)
	assert.Equal(t, "http", protocol.DefValue)

	// --wait stays opt-in (intentional CLI UX) but must note the SDK default.
	waitFlag := flags.Lookup("wait")
	require.NotNil(t, waitFlag)
	assert.Equal(t, "false", waitFlag.DefValue, "--wait must stay opt-in")
	assert.Contains(t, waitFlag.Usage, "SDKs wait by default")

	cleanupFlag := flags.Lookup("cleanup-on-failure")
	require.NotNil(t, cleanupFlag)
	assert.Equal(t, "true", cleanupFlag.DefValue)

	pollFlag := flags.Lookup("poll-interval")
	require.NotNil(t, pollFlag)
	assert.Equal(t, "0.5", pollFlag.DefValue)
}

func TestParseSandboxDefinitionFlags_EnableMeshTriState(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want koyeb.DeploymentMesh
	}{
		{"unset stays AUTO", nil, koyeb.DEPLOYMENTMESH_AUTO},
		{"set enables mesh", []string{"--enable-mesh"}, koyeb.DEPLOYMENTMESH_ENABLED},
		{"explicit false disables mesh", []string{"--enable-mesh=false"}, koyeb.DEPLOYMENTMESH_DISABLED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := parseSandboxDefinition(t, tt.args)
			require.NoError(t, err)
			assert.Equal(t, tt.want, def.GetMesh())
		})
	}
}

func TestConfigureSandboxPortsAndRoutes_ExposedPortProtocol(t *testing.T) {
	def := koyeb.NewDeploymentDefinitionWithDefaults()
	configureSandboxPortsAndRoutes(def, "http2")

	ports := def.GetPorts()
	require.Len(t, ports, 2)
	assert.Equal(t, int64(3030), ports[0].GetPort())
	assert.Equal(t, "http", ports[0].GetProtocol())
	assert.Equal(t, int64(3031), ports[1].GetPort())
	assert.Equal(t, "http2", ports[1].GetProtocol())
}

func TestParseSandboxDefinitionFlags_ExposedPortProtocolValidation(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"ftp", "ftp"},
		{"tcp", "tcp"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseSandboxDefinition(t, []string{"--exposed-port-protocol", tt.value})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Invalid protocol '"+tt.value+"'. Must be one of ('http', 'http2')")
		})
	}
}

func TestParseSandboxDefinitionFlags_TCPTProxy(t *testing.T) {
	t.Run("enabled sets proxy port 3031 tcp", func(t *testing.T) {
		def, err := parseSandboxDefinition(t, []string{"--enable-tcp-proxy"})
		require.NoError(t, err)

		proxyPorts := def.GetProxyPorts()
		require.Len(t, proxyPorts, 1)
		assert.Equal(t, int64(3031), proxyPorts[0].GetPort())
		assert.Equal(t, koyeb.PROXYPORTPROTOCOL_TCP, proxyPorts[0].GetProtocol())
	})

	t.Run("unset sets no proxy ports", func(t *testing.T) {
		def, err := parseSandboxDefinition(t, nil)
		require.NoError(t, err)
		assert.False(t, def.HasProxyPorts())
	})
}

func TestSandboxCreateInstanceTypeDefaultsToMicro(t *testing.T) {
	def, err := parseSandboxDefinition(t, nil)
	require.NoError(t, err)

	instanceTypes := def.GetInstanceTypes()
	require.Len(t, instanceTypes, 1)
	assert.Equal(t, "micro", instanceTypes[0].GetType())
}

func parseSandboxSecret(t *testing.T, args []string) (*cobra.Command, *koyeb.DeploymentDefinition) {
	t.Helper()
	cmd := sandboxCreateCmd(t)
	require.NoError(t, cmd.Flags().Parse(args))
	def := koyeb.NewDeploymentDefinitionWithDefaults()
	require.NoError(t, parseSandboxDefinitionFlags(&CLIContext{}, cmd, def, NewServiceHandler()))
	return cmd, def
}

func TestApplySandboxSecretFlag(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantEnv string
	}{
		{
			name:    "explicit flag wins over --env",
			args:    []string{"--sandbox-secret", "flag-secret", "--env", "SANDBOX_SECRET=env-secret"},
			wantEnv: "flag-secret",
		},
		{
			name:    "flag without --env",
			args:    []string{"--sandbox-secret", "flag-secret"},
			wantEnv: "flag-secret",
		},
		{
			name:    "no flag keeps --env value",
			args:    []string{"--env", "SANDBOX_SECRET=env-secret"},
			wantEnv: "env-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, def := parseSandboxSecret(t, tt.args)

			applySandboxSecretFlag(cmd.Flags(), def)
			ensureSandboxSecret(def)

			var found []string
			for _, env := range def.GetEnv() {
				if env.GetKey() == SandboxSecretKey {
					found = append(found, env.GetValue())
				}
			}
			require.Len(t, found, 1, "SANDBOX_SECRET must appear exactly once")
			assert.Equal(t, tt.wantEnv, found[0])
		})
	}
}

func sandboxSecretValues(def *koyeb.DeploymentDefinition) []string {
	var found []string
	for _, env := range def.GetEnv() {
		if env.GetKey() == SandboxSecretKey {
			found = append(found, env.GetValue())
		}
	}
	return found
}

func TestEnsureSandboxSecretGeneratesWhenNoFlagAndNoEnv(t *testing.T) {
	_, def := parseSandboxSecret(t, nil)
	ensureSandboxSecret(def)

	found := sandboxSecretValues(def)
	require.Len(t, found, 1)
	// 32 random bytes, URL-safe base64: 43 chars
	assert.Len(t, found[0], 43)
}

func TestEnsureSandboxSecretGenerationsAreUnique(t *testing.T) {
	first := koyeb.NewDeploymentDefinitionWithDefaults()
	ensureSandboxSecret(first)
	second := koyeb.NewDeploymentDefinitionWithDefaults()
	ensureSandboxSecret(second)

	assert.NotEqual(t, sandboxSecretValues(first)[0], sandboxSecretValues(second)[0])
}

func instanceSnapshot(id, name string, snapshotType koyeb.InstanceSnapshotType) *koyeb.InstanceSnapshot {
	return &koyeb.InstanceSnapshot{Id: &id, Name: &name, Type: &snapshotType}
}

func TestResolveSnapshotRef(t *testing.T) {
	filesystem := koyeb.INSTANCESNAPSHOTTYPE_FILESYSTEM
	full := koyeb.INSTANCESNAPSHOTTYPE_FULL
	getID := "323e4567-e89b-42d3-a456-426614174000"

	tests := []struct {
		name     string
		ref      string
		get      func(context.Context, string) (*koyeb.InstanceSnapshot, error)
		list     func(context.Context) ([]koyeb.InstanceSnapshot, error)
		wantID   string
		wantType koyeb.InstanceSnapshotType
	}{
		{
			name: "id lookup succeeds",
			ref:  "323e4567-e89b-42d3-a456-426614174000",
			get: func(_ context.Context, id string) (*koyeb.InstanceSnapshot, error) {
				return instanceSnapshot(id, "ignored", full), nil
			},
			list: func(context.Context) ([]koyeb.InstanceSnapshot, error) {
				t.Error("list must not be called when the id lookup succeeds")
				return nil, nil
			},
			wantID:   getID,
			wantType: full,
		},
		{
			name: "id lookup fails, name matches",
			ref:  "golden-image",
			get: func(context.Context, string) (*koyeb.InstanceSnapshot, error) {
				return nil, fmt.Errorf("not found")
			},
			list: func(context.Context) ([]koyeb.InstanceSnapshot, error) {
				return []koyeb.InstanceSnapshot{*instanceSnapshot(getID, "golden-image", filesystem)}, nil
			},
			wantID:   getID,
			wantType: filesystem,
		},
		{
			name: "id and name lookups fail, ref used as id",
			ref:  "323e4567-e89b-42d3-a456-426614174000",
			get: func(context.Context, string) (*koyeb.InstanceSnapshot, error) {
				return nil, fmt.Errorf("boom")
			},
			list: func(context.Context) ([]koyeb.InstanceSnapshot, error) {
				return nil, fmt.Errorf("boom")
			},
			wantID:   "323e4567-e89b-42d3-a456-426614174000",
			wantType: koyeb.INSTANCESNAPSHOTTYPE_FILESYSTEM,
		},
		{
			name: "name list has no exact match, ref used as id",
			ref:  "golden",
			get: func(context.Context, string) (*koyeb.InstanceSnapshot, error) {
				return nil, fmt.Errorf("not found")
			},
			list: func(context.Context) ([]koyeb.InstanceSnapshot, error) {
				return []koyeb.InstanceSnapshot{*instanceSnapshot(getID, "golden-image", filesystem)}, nil
			},
			wantID:   "golden",
			wantType: koyeb.INSTANCESNAPSHOTTYPE_FILESYSTEM,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, snapshotType := resolveSnapshotRef(context.Background(), tt.ref, tt.get, tt.list)
			assert.Equal(t, tt.wantID, id)
			assert.Equal(t, tt.wantType, snapshotType)
		})
	}
}

func TestWireSnapshot(t *testing.T) {
	filesystem := koyeb.INSTANCESNAPSHOTTYPE_FILESYSTEM
	full := koyeb.INSTANCESNAPSHOTTYPE_FULL
	snapshotID := "323e4567-e89b-42d3-a456-426614174000"

	t.Run("filesystem snapshot keeps definition", func(t *testing.T) {
		createService := koyeb.NewCreateServiceWithDefaults()
		createService.SetDefinition(*koyeb.NewDeploymentDefinitionWithDefaults())

		wireSnapshot(createService, snapshotID, filesystem, "my-sandbox")

		assert.Equal(t, snapshotID, createService.GetInstanceSnapshotId())
		assert.True(t, createService.HasDefinition())
		assert.False(t, createService.HasName(), "name stays on the definition")
	})

	t.Run("full snapshot drops definition and names the service", func(t *testing.T) {
		createService := koyeb.NewCreateServiceWithDefaults()
		createService.SetDefinition(*koyeb.NewDeploymentDefinitionWithDefaults())

		wireSnapshot(createService, snapshotID, full, "my-sandbox")

		assert.Equal(t, snapshotID, createService.GetInstanceSnapshotId())
		assert.False(t, createService.HasDefinition(), "the API infers a FULL snapshot's definition")
		assert.Equal(t, "my-sandbox", createService.GetName())
	})
}

func TestWaitPollInterval(t *testing.T) {
	t.Run("sandbox create defaults to 0.5s", func(t *testing.T) {
		assert.Equal(t, 500*time.Millisecond, waitPollInterval(sandboxCreateCmd(t)))
	})

	t.Run("sandbox create honors --poll-interval", func(t *testing.T) {
		cmd := sandboxCreateCmd(t)
		require.NoError(t, cmd.Flags().Set("poll-interval", "1.5"))
		assert.Equal(t, 1500*time.Millisecond, waitPollInterval(cmd))
	})

	t.Run("non-positive poll interval falls back to the flag default", func(t *testing.T) {
		cmd := sandboxCreateCmd(t)
		require.NoError(t, cmd.Flags().Set("poll-interval", "0"))
		assert.Equal(t, 500*time.Millisecond, waitPollInterval(cmd))
	})

	t.Run("non-finite poll intervals fall back to the flag default", func(t *testing.T) {
		for _, value := range []string{"nan", "inf", "-inf"} {
			cmd := sandboxCreateCmd(t)
			require.NoError(t, cmd.Flags().Set("poll-interval", value))
			assert.Equal(t, 500*time.Millisecond, waitPollInterval(cmd), "--poll-interval %s must not reach time.NewTicker", value)
		}
	})

	t.Run("commands without the flag poll at 2s", func(t *testing.T) {
		cmd, _, err := NewServiceCmd().Find([]string{"create"})
		require.NoError(t, err)
		assert.Equal(t, 2*time.Second, waitPollInterval(cmd))
	})
}

func TestDeploymentWaitDone(t *testing.T) {
	tests := []struct {
		status koyeb.ServiceStatus
		done   bool
		failed bool
	}{
		{koyeb.SERVICESTATUS_HEALTHY, true, false},
		{koyeb.SERVICESTATUS_PAUSED, true, false},
		{koyeb.SERVICESTATUS_DELETED, true, true},
		{koyeb.SERVICESTATUS_DEGRADED, true, true},
		{koyeb.SERVICESTATUS_UNHEALTHY, true, true},
		{koyeb.SERVICESTATUS_STARTING, false, false},
		{koyeb.SERVICESTATUS_RESUMING, false, false},
		{koyeb.SERVICESTATUS_DELETING, false, false},
		{koyeb.SERVICESTATUS_PAUSING, false, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			done, failed := deploymentWaitDone(tt.status)
			assert.Equal(t, tt.done, done)
			assert.Equal(t, tt.failed, failed)
		})
	}
}

func TestPoolCreateInstanceTypeDefaultsToMicro(t *testing.T) {
	cmd := newPoolCreateCmd()

	req, err := buildCreateServicePool(&CLIContext{}, cmd, "micro-pool")
	require.NoError(t, err)

	def := req.GetDefinition()
	instanceTypes := def.GetInstanceTypes()
	require.Len(t, instanceTypes, 1)
	assert.Equal(t, "micro", instanceTypes[0].GetType())
}

// fakeSandboxCreate captures the API seams of the create flow so the
// wiring — snapshot, protocol, secret ordering, cleanup — is pinned
// without a live client.
type fakeSandboxCreate struct {
	existingAppID string
	createdAppID  string
	snapshotID    string
	snapshotType  koyeb.InstanceSnapshotType
	serviceID     string
	createErr     error
	waitErr       error

	createdApps     []string
	deletedApps     []string
	deletedServices []string
	rendered        []string
	createReq       *koyeb.CreateService
}

func (f *fakeSandboxCreate) deps() sandboxCreateDeps {
	return sandboxCreateDeps{
		getAppID: func(*CLIContext, string) (string, error) { return f.existingAppID, nil },
		createApp: func(_ *CLIContext, name string) (string, error) {
			f.createdApps = append(f.createdApps, name)
			return f.createdAppID, nil
		},
		resolveSnapshot: func(*CLIContext, string) (string, koyeb.InstanceSnapshotType) {
			return f.snapshotID, f.snapshotType
		},
		createService: func(_ *CLIContext, _ *cobra.Command, _ []string, req *koyeb.CreateService) (*koyeb.Service, error) {
			f.createReq = req
			if f.createErr != nil {
				return nil, f.createErr
			}
			id := f.serviceID
			return &koyeb.Service{Id: &id}, nil
		},
		waitForService: func(*CLIContext, *cobra.Command, string) error { return f.waitErr },
		deleteApp: func(_ *CLIContext, appID string) {
			f.deletedApps = append(f.deletedApps, appID)
		},
		deleteService: func(_ *CLIContext, serviceID string) {
			f.deletedServices = append(f.deletedServices, serviceID)
		},
		renderService: func(_ *CLIContext, _ *cobra.Command, serviceID string) {
			f.rendered = append(f.rendered, serviceID)
		},
	}
}

func TestCreateSandboxCleansUpAutoCreatedAppOnCreateFailure(t *testing.T) {
	fake := &fakeSandboxCreate{createdAppID: "app-123", serviceID: "svc-123", createErr: fmt.Errorf("api down")}
	cmd := sandboxCreateCmd(t)

	err := createSandbox(&CLIContext{}, cmd, []string{"myapp/mysbx"}, fake.deps())
	require.Error(t, err)
	assert.Equal(t, []string{"myapp"}, fake.createdApps, "missing app must be auto-created")
	assert.Equal(t, []string{"app-123"}, fake.deletedApps, "auto-created app must be deleted on create failure")
	assert.Empty(t, fake.deletedServices)
	assert.Empty(t, fake.rendered)
}

func TestCreateSandboxCleansUpAutoCreatedAppOnFlagValidationFailure(t *testing.T) {
	fake := &fakeSandboxCreate{createdAppID: "app-123", serviceID: "svc-123"}
	cmd := sandboxCreateCmd(t)
	require.NoError(t, cmd.Flags().Set("exposed-port-protocol", "ftp"))

	err := createSandbox(&CLIContext{}, cmd, []string{"myapp/mysbx"}, fake.deps())
	require.Error(t, err)
	assert.Equal(t, []string{"app-123"}, fake.deletedApps, "validation failure after app creation must clean the app up")
	assert.Nil(t, fake.createReq, "no service create request may be sent")
}

func TestCreateSandboxKeepsExistingAppOnCreateFailure(t *testing.T) {
	fake := &fakeSandboxCreate{existingAppID: "app-existing", serviceID: "svc-123", createErr: fmt.Errorf("api down")}
	cmd := sandboxCreateCmd(t)

	err := createSandbox(&CLIContext{}, cmd, []string{"myapp/mysbx"}, fake.deps())
	require.Error(t, err)
	assert.Empty(t, fake.createdApps, "existing app must not be re-created")
	assert.Empty(t, fake.deletedApps, "an app this command did not create must never be deleted")
}

func TestCreateSandboxServiceCleanupOnWaitFailure(t *testing.T) {
	fake := &fakeSandboxCreate{existingAppID: "app-existing", serviceID: "svc-123", waitErr: fmt.Errorf("timed out")}
	cmd := sandboxCreateCmd(t)
	require.NoError(t, cmd.Flags().Set("wait", "true"))

	err := createSandbox(&CLIContext{}, cmd, []string{"myapp/mysbx"}, fake.deps())
	require.Error(t, err)
	assert.Equal(t, []string{"svc-123"}, fake.deletedServices, "default cleanup deletes the sandbox on wait failure")
	assert.Empty(t, fake.rendered, "a failed sandbox is not rendered after cleanup")
}

func TestCreateSandboxNoServiceCleanupWhenFlagDisabled(t *testing.T) {
	fake := &fakeSandboxCreate{existingAppID: "app-existing", serviceID: "svc-123", waitErr: fmt.Errorf("timed out")}
	cmd := sandboxCreateCmd(t)
	require.NoError(t, cmd.Flags().Set("wait", "true"))
	require.NoError(t, cmd.Flags().Set("cleanup-on-failure", "false"))

	err := createSandbox(&CLIContext{}, cmd, []string{"myapp/mysbx"}, fake.deps())
	require.Error(t, err)
	assert.Empty(t, fake.deletedServices)
}

func TestCreateSandboxRendersOnSuccess(t *testing.T) {
	fake := &fakeSandboxCreate{existingAppID: "app-existing", serviceID: "svc-123"}
	cmd := sandboxCreateCmd(t)

	err := createSandbox(&CLIContext{}, cmd, []string{"myapp/mysbx"}, fake.deps())
	require.NoError(t, err)
	assert.Equal(t, []string{"svc-123"}, fake.rendered)
	assert.Empty(t, fake.deletedApps)
	assert.Empty(t, fake.deletedServices)
}

func TestCreateSandboxWiresFullSnapshot(t *testing.T) {
	fake := &fakeSandboxCreate{
		existingAppID: "app-existing",
		serviceID:     "svc-123",
		snapshotID:    "snap-123",
		snapshotType:  koyeb.INSTANCESNAPSHOTTYPE_FULL,
	}
	cmd := sandboxCreateCmd(t)

	require.NoError(t, createSandbox(&CLIContext{}, cmd, []string{"myapp/mysbx"}, fake.deps()))

	req := fake.createReq
	require.NotNil(t, req)
	assert.Equal(t, "snap-123", req.GetInstanceSnapshotId())
	assert.False(t, req.HasDefinition(), "FULL snapshot: the API infers the definition")
	assert.Equal(t, "mysbx", req.GetName())
}

func TestCreateSandboxWiresExposedPortProtocol(t *testing.T) {
	fake := &fakeSandboxCreate{existingAppID: "app-existing", serviceID: "svc-123"}
	cmd := sandboxCreateCmd(t)
	require.NoError(t, cmd.Flags().Set("exposed-port-protocol", "http2"))

	require.NoError(t, createSandbox(&CLIContext{}, cmd, []string{"myapp/mysbx"}, fake.deps()))

	def := fake.createReq.GetDefinition()
	ports := def.GetPorts()
	require.Len(t, ports, 2)
	assert.Equal(t, "http", ports[0].GetProtocol())
	assert.Equal(t, "http2", ports[1].GetProtocol())
}

func TestCreateSandboxWiresSandboxSecret(t *testing.T) {
	fake := &fakeSandboxCreate{existingAppID: "app-existing", serviceID: "svc-123"}
	cmd := sandboxCreateCmd(t)
	require.NoError(t, cmd.Flags().Set("sandbox-secret", "flag-secret"))
	require.NoError(t, cmd.Flags().Set("env", "SANDBOX_SECRET=env-secret"))

	require.NoError(t, createSandbox(&CLIContext{}, cmd, []string{"myapp/mysbx"}, fake.deps()))

	secrets := []string{}
	def := fake.createReq.GetDefinition()
	for _, env := range def.GetEnv() {
		if env.GetKey() == SandboxSecretKey {
			secrets = append(secrets, env.GetValue())
		}
	}
	require.Len(t, secrets, 1)
	assert.Equal(t, "flag-secret", secrets[0], "--sandbox-secret must win over --env")
}

func TestWaitTimeoutFlagRejectsNonPositiveValues(t *testing.T) {
	cmd := sandboxCreateCmd(t)
	require.NoError(t, cmd.Flags().Set("wait-timeout", "0"))

	_, err := waitTimeoutFlag(cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "positive")

	require.NoError(t, cmd.Flags().Set("wait-timeout", "1m"))
	timeout, err := waitTimeoutFlag(cmd)
	require.NoError(t, err)
	assert.Equal(t, time.Minute, timeout)
}
