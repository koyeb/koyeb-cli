package koyeb

import (
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The instance-type/regions flag bundle is the single registration and
// parse seam shared by the service, sandbox and pool surfaces. These
// tests pin the bundle contract; the per-surface suites pin the merged
// behavior.

func TestAddInstanceTypeRegionsFlagsServiceSurface(t *testing.T) {
	flags := pflag.NewFlagSet("service", pflag.ContinueOnError)
	addInstanceTypeRegionsFlags(flags, serviceInstanceTypeRegionsFlagUsage)

	expected := []struct {
		name     string
		defValue string
		usage    string
	}{
		{
			"instance-type", "nano", "Instance type",
		},
		{
			"regions", "[]",
			"Add a region where the service is deployed. " +
				"You can specify this flag multiple times to deploy the service in multiple regions.\n" +
				"To update a service and remove a region, prefix the region name with '!', for example --region '!par'\n" +
				"If the region is not specified on service creation, the service is deployed in was\n",
		},
	}
	for _, want := range expected {
		flag := flags.Lookup(want.name)
		require.NotNil(t, flag, "--%s must be registered by the instance-type/regions bundle", want.name)
		assert.Equal(t, want.defValue, flag.DefValue, "--%s default value", want.name)
		assert.Equal(t, want.usage, flag.Usage, "--%s help text", want.name)
	}
}

func TestAddInstanceTypeRegionsFlagsSandboxPoolSurface(t *testing.T) {
	flags := pflag.NewFlagSet("sandbox-pool", pflag.ContinueOnError)
	addInstanceTypeRegionsFlags(flags, sandboxPoolInstanceTypeRegionsFlagUsage)

	expected := []struct {
		name     string
		defValue string
		usage    string
	}{
		{"instance-type", "micro", "Instance type"},
		{"regions", "[]", "Deployment regions"},
	}
	for _, want := range expected {
		flag := flags.Lookup(want.name)
		require.NotNil(t, flag, "--%s must be registered by the instance-type/regions bundle", want.name)
		assert.Equal(t, want.defValue, flag.DefValue, "--%s default value", want.name)
		assert.Equal(t, want.usage, flag.Usage, "--%s help text", want.name)
	}
}

// instanceTypeRegionsFlagNames returns the flag names the bundle
// registers for the given skin — the bundle union the surface flag-set
// tests derive their expectations from.
func instanceTypeRegionsFlagNames(usage instanceTypeRegionsFlagUsage) []string {
	flags := pflag.NewFlagSet("instance-type-regions", pflag.ContinueOnError)
	addInstanceTypeRegionsFlags(flags, usage)
	names := make([]string, 0, 2)
	flags.VisitAll(func(f *pflag.Flag) { names = append(names, f.Name) })
	return names
}

func instanceTypeRegionsTestFlagSet(usage instanceTypeRegionsFlagUsage) *pflag.FlagSet {
	flags := pflag.NewFlagSet("instance-type-regions", pflag.ContinueOnError)
	addInstanceTypeRegionsFlags(flags, usage)
	return flags
}

// liveInstanceTypeRegionsDefinition returns a definition carrying a live
// instance type and regions, so the changed-only merge conventions are
// observable.
func liveInstanceTypeRegionsDefinition() *koyeb.DeploymentDefinition {
	def := koyeb.NewDeploymentDefinitionWithDefaults()

	instanceType := koyeb.NewDeploymentInstanceTypeWithDefaults()
	instanceType.SetType("nano")
	def.SetInstanceTypes([]koyeb.DeploymentInstanceType{*instanceType})

	def.SetRegions([]string{"par", "fra"})
	return def
}

func TestParseInstanceTypeRegionsNoFlagsKeepsLiveValues(t *testing.T) {
	h := &ServiceHandler{}
	flags := instanceTypeRegionsTestFlagSet(serviceInstanceTypeRegionsFlagUsage)

	def := liveInstanceTypeRegionsDefinition()
	require.NoError(t, h.parseInstanceTypeRegions(flags, def))

	instanceTypes := def.GetInstanceTypes()
	require.Len(t, instanceTypes, 1)
	assert.Equal(t, "nano", instanceTypes[0].GetType())
	assert.Equal(t, []string{"par", "fra"}, def.GetRegions())
}

func TestParseInstanceTypeRegionsChangedFlagsOverrideLiveValues(t *testing.T) {
	h := &ServiceHandler{}
	flags := instanceTypeRegionsTestFlagSet(serviceInstanceTypeRegionsFlagUsage)
	require.NoError(t, flags.Parse([]string{"--instance-type", "micro", "--regions", "was"}))

	def := liveInstanceTypeRegionsDefinition()
	require.NoError(t, h.parseInstanceTypeRegions(flags, def))

	// A changed --instance-type replaces the live list with one entry.
	instanceTypes := def.GetInstanceTypes()
	require.Len(t, instanceTypes, 1)
	assert.Equal(t, "micro", instanceTypes[0].GetType())

	// The regions merge: a new region is appended to the live ones.
	assert.Equal(t, []string{"par", "fra", "was"}, def.GetRegions())
}

func TestParseInstanceTypeRegionsDeletionIdiom(t *testing.T) {
	h := &ServiceHandler{}
	flags := instanceTypeRegionsTestFlagSet(serviceInstanceTypeRegionsFlagUsage)
	require.NoError(t, flags.Parse([]string{"--regions", "!par"}))

	def := liveInstanceTypeRegionsDefinition()
	require.NoError(t, h.parseInstanceTypeRegions(flags, def))

	assert.Equal(t, []string{"fra"}, def.GetRegions(), "the deletion idiom must remove the live region")
}

func TestParseInstanceTypeRegionsFreshDefinitionDefaults(t *testing.T) {
	h := &ServiceHandler{}

	// The instance-type default is per-surface: parseInstanceType reads it
	// back from the registered flag, so the skin pins what a fresh
	// definition gets when --instance-type is not passed.
	t.Run("service skin defaults to nano and was", func(t *testing.T) {
		flags := instanceTypeRegionsTestFlagSet(serviceInstanceTypeRegionsFlagUsage)

		def := koyeb.NewDeploymentDefinitionWithDefaults()
		require.NoError(t, h.parseInstanceTypeRegions(flags, def))

		instanceTypes := def.GetInstanceTypes()
		require.Len(t, instanceTypes, 1)
		assert.Equal(t, "nano", instanceTypes[0].GetType())
		assert.Equal(t, []string{"was"}, def.GetRegions())
	})

	t.Run("sandbox/pool skin defaults to micro and was", func(t *testing.T) {
		flags := instanceTypeRegionsTestFlagSet(sandboxPoolInstanceTypeRegionsFlagUsage)

		def := koyeb.NewDeploymentDefinitionWithDefaults()
		require.NoError(t, h.parseInstanceTypeRegions(flags, def))

		instanceTypes := def.GetInstanceTypes()
		require.Len(t, instanceTypes, 1)
		assert.Equal(t, "micro", instanceTypes[0].GetType())
		assert.Equal(t, []string{"was"}, def.GetRegions())
	})
}
