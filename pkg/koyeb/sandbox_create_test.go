package koyeb

import (
	"context"
	"fmt"
	"testing"

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

func TestParseSandboxDefinitionFlags_ExposedPortProtocol(t *testing.T) {
	_, err := parseSandboxDefinition(t, []string{"--exposed-port-protocol", "http2"})
	require.NoError(t, err)
}

func TestConfigureSandboxPortsAndRoutes_ExposedPortProtocol(t *testing.T) {
	def := koyeb.NewDeploymentDefinitionWithDefaults()
	configureSandboxPortsAndRoutes(def, "http2", false, false)

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

func TestApplySandboxSecretFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantEnv  map[string]string
		generated bool
	}{
		{
			name:     "explicit flag wins over --env",
			args:     []string{"--sandbox-secret", "flag-secret", "--env", "SANDBOX_SECRET=env-secret"},
			wantEnv:  map[string]string{"SANDBOX_SECRET": "flag-secret"},
		},
		{
			name:     "flag without --env",
			args:     []string{"--sandbox-secret", "flag-secret"},
			wantEnv:  map[string]string{"SANDBOX_SECRET": "flag-secret"},
		},
		{
			name:     "no flag keeps --env value",
			args:     []string{"--env", "SANDBOX_SECRET=env-secret"},
			wantEnv:  map[string]string{"SANDBOX_SECRET": "env-secret"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := sandboxCreateCmd(t)
			require.NoError(t, cmd.Flags().Parse(tt.args))

			def, err := parseSandboxDefinition(t, tt.args)
			require.NoError(t, err)

			applySandboxSecretFlag(cmd.Flags(), def)
			ensureSandboxSecret(def)

			var found []string
			for _, env := range def.GetEnv() {
				if env.GetKey() == SandboxSecretKey {
					found = append(found, env.GetValue())
				}
			}
			require.Len(t, found, 1, "SANDBOX_SECRET must appear exactly once")
			assert.Equal(t, tt.wantEnv[SandboxSecretKey], found[0])
		})
	}
}

func TestEnsureSandboxSecretGeneratesWhenNoFlagAndNoEnv(t *testing.T) {
	def, err := parseSandboxDefinition(t, nil)
	require.NoError(t, err)

	applySandboxSecretFlag(sandboxCreateCmd(t).Flags(), def)
	ensureSandboxSecret(def)

	var found []string
	for _, env := range def.GetEnv() {
		if env.GetKey() == SandboxSecretKey {
			found = append(found, env.GetValue())
		}
	}
	require.Len(t, found, 1)
	// 32 random bytes, URL-safe base64: 43 chars
	assert.Len(t, found[0], 43)
}

func instanceSnapshot(id, name string, snapshotType koyeb.InstanceSnapshotType) *koyeb.InstanceSnapshot {
	return &koyeb.InstanceSnapshot{Id: &id, Name: &name, Type: &snapshotType}
}

func TestResolveSnapshotRef(t *testing.T) {
	filesystem := koyeb.INSTANCESNAPSHOTTYPE_FILESYSTEM
	full := koyeb.INSTANCESNAPSHOTTYPE_FULL
	getID := "323e4567-e89b-42d3-a456-426614174000"

	tests := []struct {
		name      string
		ref       string
		get       func(context.Context, string) (*koyeb.InstanceSnapshot, error)
		list      func(context.Context) ([]koyeb.InstanceSnapshot, error)
		wantID    string
		wantType  koyeb.InstanceSnapshotType
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

func TestPoolCreateInstanceTypeDefaultsToMicro(t *testing.T) {
	cmd := newPoolCreateCmd()

	req, err := buildCreateServicePool(&CLIContext{}, cmd, "micro-pool")
	require.NoError(t, err)

	def := req.GetDefinition()
	instanceTypes := def.GetInstanceTypes()
	require.Len(t, instanceTypes, 1)
	assert.Equal(t, "micro", instanceTypes[0].GetType())
}
