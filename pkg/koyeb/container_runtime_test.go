package koyeb

import (
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// containerRuntimeFlagNames returns the flag names the container-runtime
// bundle registers — the bundle union the surface flag-set tests derive
// their expectations from.
func containerRuntimeFlagNames() []string {
	flags := pflag.NewFlagSet("container-runtime", pflag.ContinueOnError)
	addContainerRuntimeFlags(flags)
	names := make([]string, 0, 1)
	flags.VisitAll(func(f *pflag.Flag) { names = append(names, f.Name) })
	return names
}

// parseServiceContainerRuntime runs the service parse path over the given
// arguments and returns the resulting definition.
func parseServiceContainerRuntime(t *testing.T, args []string, live *koyeb.DeploymentDefinition) *koyeb.DeploymentDefinition {
	t.Helper()
	cmd := serviceCreateCmdForTest(t)
	require.NoError(t, cmd.Flags().Parse(args))
	definition := live
	if definition == nil {
		definition = koyeb.NewDeploymentDefinitionWithDefaults()
	}
	require.NoError(t, NewServiceHandler().parseServiceDefinitionFlags(&CLIContext{}, cmd.Flags(), definition))
	return definition
}

func TestApplyContainerRuntimeFlagServices(t *testing.T) {
	t.Run("unset leaves the definition empty (region default)", func(t *testing.T) {
		def := parseServiceContainerRuntime(t, nil, nil)
		assert.False(t, def.HasContainerRuntime(), "the runtime must stay unset for the region default")
	})

	t.Run("--container-runtime kata reads back as kata", func(t *testing.T) {
		def := parseServiceContainerRuntime(t, []string{"--container-runtime", "kata"}, nil)
		require.True(t, def.HasContainerRuntime())
		assert.Equal(t, "kata", def.GetContainerRuntime())
	})

	t.Run("unset flag keeps the stored runtime on update", func(t *testing.T) {
		live := koyeb.NewDeploymentDefinitionWithDefaults()
		live.SetContainerRuntime("kata")

		def := parseServiceContainerRuntime(t, nil, live)
		require.True(t, def.HasContainerRuntime())
		assert.Equal(t, "kata", def.GetContainerRuntime())
	})

	t.Run("an explicit empty value drops back to the region default", func(t *testing.T) {
		live := koyeb.NewDeploymentDefinitionWithDefaults()
		live.SetContainerRuntime("kata")

		def := parseServiceContainerRuntime(t, []string{"--container-runtime", ""}, live)
		assert.False(t, def.HasContainerRuntime(), "an explicit empty value must drop to the region default")
	})

	// The CLI does not validate the allowed runtimes: the API rejects an
	// unknown value and its message must surface unchanged. The flag
	// value reaches the definition verbatim.
	t.Run("an unknown value is sent verbatim for the API to reject", func(t *testing.T) {
		def := parseServiceContainerRuntime(t, []string{"--container-runtime", "hyperdrive"}, nil)
		require.True(t, def.HasContainerRuntime())
		assert.Equal(t, "hyperdrive", def.GetContainerRuntime())
	})
}
