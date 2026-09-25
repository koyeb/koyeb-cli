package koyeb

import (
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

func TestPoolCreateInstanceTypeDefaultsToMicro(t *testing.T) {
	cmd := newPoolCreateCmd()

	req, err := buildCreateServicePool(&CLIContext{}, cmd, "micro-pool")
	require.NoError(t, err)

	def := req.GetDefinition()
	instanceTypes := def.GetInstanceTypes()
	require.Len(t, instanceTypes, 1)
	assert.Equal(t, "micro", instanceTypes[0].GetType())
}
