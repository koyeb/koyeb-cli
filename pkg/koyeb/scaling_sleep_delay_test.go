package koyeb

import (
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The scaling/sleep-delay flag bundle is the single registration and
// parse seam shared by the service, sandbox and pool surfaces. These
// tests pin the bundle contract; the per-surface suites pin the merged
// behavior.

func TestAddScalingSleepDelayFlagsServiceSurface(t *testing.T) {
	flags := pflag.NewFlagSet("service", pflag.ContinueOnError)
	addScalingSleepDelayFlags(flags, serviceScalingSleepDelayFlagUsage)

	expected := []struct {
		name     string
		defValue string
		usage    string
	}{
		{"scale", "1", "Set both min-scale and max-scale"},
		{"min-scale", "1", "Min scale"},
		{"max-scale", "1", "Max scale"},
		{
			"autoscaling-average-cpu", "0",
			"Target CPU usage (in %) to trigger a scaling event. Set to 0 to disable CPU autoscaling.",
		},
		{
			"autoscaling-average-mem", "0",
			"Target memory usage (in %) to trigger a scaling event. Set to 0 to disable memory autoscaling.",
		},
		{
			"autoscaling-requests-per-second", "0",
			"Target requests per second to trigger a scaling event. Set to 0 to disable requests per second autoscaling.",
		},
		{
			"autoscaling-concurrent-requests", "0",
			"Target concurrent requests to trigger a scaling event. Set to 0 to disable concurrent requests autoscaling.",
		},
		{
			"autoscaling-requests-response-time", "0",
			"Target p95 response time to trigger a scaling event (in ms). " +
				"Set to 0 to disable concurrent response time autoscaling.",
		},
		{
			"light-sleep-delay", "0s",
			"Delay after which an idle service is put to light sleep. " +
				"Use duration format (e.g., '1m', '5m', '1h'). Set to 0 to disable.",
		},
		{
			"deep-sleep-delay", "0s",
			"Delay after which an idle service is put to deep sleep. " +
				"Use duration format (e.g., '5m', '30m', '1h'). Set to 0 to disable.",
		},
	}
	for _, want := range expected {
		flag := flags.Lookup(want.name)
		require.NotNil(t, flag, "--%s must be registered by the scaling/sleep-delay bundle", want.name)
		assert.Equal(t, want.defValue, flag.DefValue, "--%s default value", want.name)
		assert.Equal(t, want.usage, flag.Usage, "--%s help text", want.name)
	}
}

func TestAddScalingSleepDelayFlagsSandboxPoolSurface(t *testing.T) {
	flags := pflag.NewFlagSet("sandbox-pool", pflag.ContinueOnError)
	addScalingSleepDelayFlags(flags, sandboxPoolScalingSleepDelayFlagUsage)

	expected := []struct {
		name     string
		defValue string
		usage    string
	}{
		{"min-scale", "1", "Min scale"},
		{
			"light-sleep-delay", "0s",
			"Delay after which an idle service is put to light sleep. " +
				"Use duration format (e.g., '1m', '5m', '1h'). Set to 0 to disable.",
		},
		{
			"deep-sleep-delay", "0s",
			"Delay after which an idle service is put to deep sleep. " +
				"Use duration format (e.g., '5m', '30m', '1h'). Set to 0 to disable.",
		},
	}
	for _, want := range expected {
		flag := flags.Lookup(want.name)
		require.NotNil(t, flag, "--%s must be registered by the scaling/sleep-delay bundle", want.name)
		assert.Equal(t, want.defValue, flag.DefValue, "--%s default value", want.name)
		assert.Equal(t, want.usage, flag.Usage, "--%s help text", want.name)
	}

	// The sandbox and pool surfaces always run single-instance: the
	// service-only scaling flags must stay service-surface flags.
	for _, name := range []string{
		"scale", "max-scale",
		"autoscaling-average-cpu", "autoscaling-average-mem",
		"autoscaling-requests-per-second", "autoscaling-concurrent-requests",
		"autoscaling-requests-response-time",
	} {
		assert.Nil(t, flags.Lookup(name), "--%s must not be registered by the sandbox/pool skin", name)
	}
}

// scalingSleepDelayFlagNames returns the flag names the bundle registers
// for the given skin — the bundle union the surface flag-set tests derive
// their expectations from.
func scalingSleepDelayFlagNames(usage scalingSleepDelayFlagUsage) []string {
	flags := pflag.NewFlagSet("scaling-sleep-delay", pflag.ContinueOnError)
	addScalingSleepDelayFlags(flags, usage)
	names := make([]string, 0, 10)
	flags.VisitAll(func(f *pflag.Flag) { names = append(names, f.Name) })
	return names
}

func scalingSleepDelayTestFlagSet(usage scalingSleepDelayFlagUsage) *pflag.FlagSet {
	flags := pflag.NewFlagSet("scaling-sleep-delay", pflag.ContinueOnError)
	addScalingSleepDelayFlags(flags, usage)
	return flags
}

// liveScalingServiceDefinition returns a definition with a live non-free
// instance type and a live scaling carrying an autoscaling target, so
// the changed-only merge conventions are observable.
func liveScalingServiceDefinition() *koyeb.DeploymentDefinition {
	def := koyeb.NewDeploymentDefinitionWithDefaults()

	instanceType := koyeb.NewDeploymentInstanceTypeWithDefaults()
	instanceType.SetType("nano")
	def.SetInstanceTypes([]koyeb.DeploymentInstanceType{*instanceType})

	scaling := koyeb.NewDeploymentScalingWithDefaults()
	scaling.SetMin(0)
	scaling.SetMax(3)
	target := koyeb.NewDeploymentScalingTarget()
	cpu := koyeb.NewDeploymentScalingTargetAverageCPU()
	cpu.SetValue(80)
	target.SetAverageCpu(*cpu)
	scaling.Targets = []koyeb.DeploymentScalingTarget{*target}
	def.SetScalings([]koyeb.DeploymentScaling{*scaling})

	return def
}

// liveScalingPoolDefinition returns a definition with the live
// single-instance scaling of a pool (or sandbox), optionally carrying a
// light sleep delay.
func liveScalingPoolDefinition(lightSleepSeconds int64) *koyeb.DeploymentDefinition {
	def := koyeb.NewDeploymentDefinitionWithDefaults()

	scaling := koyeb.NewDeploymentScalingWithDefaults()
	scaling.SetMin(1)
	scaling.SetMax(1)
	if lightSleepSeconds > 0 {
		scaling.SetMin(0)
		delays := koyeb.NewDeploymentScalingTargetSleepIdleDelay()
		delays.SetLightSleepValue(lightSleepSeconds)
		target := koyeb.NewDeploymentScalingTarget()
		target.SetSleepIdleDelay(*delays)
		scaling.Targets = []koyeb.DeploymentScalingTarget{*target}
	}
	def.SetScalings([]koyeb.DeploymentScaling{*scaling})

	return def
}

func TestParseScalingSleepDelayServiceSurface(t *testing.T) {
	h := &ServiceHandler{}

	t.Run("no flags keep the live scaling", func(t *testing.T) {
		flags := scalingSleepDelayTestFlagSet(serviceScalingSleepDelayFlagUsage)

		def := liveScalingServiceDefinition()
		require.NoError(t, h.parseScalingSleepDelay(flags, def, scalingSleepDelayParseOptions{}))

		scalings := def.GetScalings()
		require.Len(t, scalings, 1)
		assert.Equal(t, int64(0), scalings[0].GetMin())
		assert.Equal(t, int64(3), scalings[0].GetMax())
		targets := scalings[0].GetTargets()
		require.Len(t, targets, 1)
		cpu := targets[0].GetAverageCpu()
		assert.Equal(t, int64(80), cpu.GetValue())
	})

	t.Run("changed flags override the live scaling", func(t *testing.T) {
		flags := scalingSleepDelayTestFlagSet(serviceScalingSleepDelayFlagUsage)
		require.NoError(t, flags.Parse([]string{"--min-scale", "2", "--autoscaling-average-cpu", "50"}))

		def := liveScalingServiceDefinition()
		require.NoError(t, h.parseScalingSleepDelay(flags, def, scalingSleepDelayParseOptions{}))

		scalings := def.GetScalings()
		require.Len(t, scalings, 1)
		assert.Equal(t, int64(2), scalings[0].GetMin())
		assert.Equal(t, int64(3), scalings[0].GetMax(), "an untouched flag keeps the live value")
		cpu := scalings[0].GetTargets()[0].GetAverageCpu()
		assert.Equal(t, int64(50), cpu.GetValue())
	})

	t.Run("sleep delays require min-scale 0", func(t *testing.T) {
		flags := scalingSleepDelayTestFlagSet(serviceScalingSleepDelayFlagUsage)
		require.NoError(t, flags.Parse([]string{"--min-scale", "1", "--light-sleep-delay", "1m"}))

		def := liveScalingServiceDefinition()
		err := h.parseScalingSleepDelay(flags, def, scalingSleepDelayParseOptions{})
		require.Error(t, err)
		assert.Contains(t, err.Error(),
			"--light-sleep-delay and --deep-sleep-delay can only be used when min-scale is 0")
	})

	t.Run("a zero sleep delay clears the live delay", func(t *testing.T) {
		flags := scalingSleepDelayTestFlagSet(serviceScalingSleepDelayFlagUsage)
		require.NoError(t, flags.Parse([]string{"--light-sleep-delay", "0"}))

		def := liveScalingPoolDefinition(300)
		require.NoError(t, h.parseScalingSleepDelay(flags, def, scalingSleepDelayParseOptions{}))

		scalings := def.GetScalings()
		require.Len(t, scalings, 1)
		assert.Empty(t, scalings[0].GetTargets(), "the cleared delay must remove its target")
	})
}

func TestParseScalingSleepDelaySingleInstanceCreateSurface(t *testing.T) {
	h := &ServiceHandler{}
	opts := scalingSleepDelayParseOptions{what: "sandbox", singleInstance: true}

	t.Run("no flags build the default single-instance scaling", func(t *testing.T) {
		flags := scalingSleepDelayTestFlagSet(sandboxPoolScalingSleepDelayFlagUsage)

		def := koyeb.NewDeploymentDefinitionWithDefaults()
		require.NoError(t, h.parseScalingSleepDelay(flags, def, opts))

		scalings := def.GetScalings()
		require.Len(t, scalings, 1)
		assert.Equal(t, int64(1), scalings[0].GetMin())
		assert.Equal(t, int64(1), scalings[0].GetMax())
		assert.Empty(t, scalings[0].GetTargets())
	})

	t.Run("min-scale 0 with a light sleep delay is accepted", func(t *testing.T) {
		flags := scalingSleepDelayTestFlagSet(sandboxPoolScalingSleepDelayFlagUsage)
		require.NoError(t, flags.Parse([]string{"--min-scale", "0", "--light-sleep-delay", "5m"}))

		def := koyeb.NewDeploymentDefinitionWithDefaults()
		require.NoError(t, h.parseScalingSleepDelay(flags, def, opts))

		scalings := def.GetScalings()
		require.Len(t, scalings, 1)
		assert.Equal(t, int64(0), scalings[0].GetMin())
		assert.Equal(t, int64(1), scalings[0].GetMax())
		targets := scalings[0].GetTargets()
		require.Len(t, targets, 1)
		delays := targets[0].GetSleepIdleDelay()
		assert.Equal(t, int64(300), delays.GetLightSleepValue())
	})

	t.Run("sleep delays require min-scale 0", func(t *testing.T) {
		flags := scalingSleepDelayTestFlagSet(sandboxPoolScalingSleepDelayFlagUsage)
		require.NoError(t, flags.Parse([]string{"--min-scale", "1", "--light-sleep-delay", "1m"}))

		def := koyeb.NewDeploymentDefinitionWithDefaults()
		err := h.parseScalingSleepDelay(flags, def, opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Error while configuring the sandbox")
		assert.Contains(t, err.Error(),
			"--light-sleep-delay and --deep-sleep-delay can only be used when min-scale is 0")
	})
}

func TestParseScalingSleepDelayPoolUpdateSurface(t *testing.T) {
	h := &ServiceHandler{}
	opts := scalingSleepDelayParseOptions{what: "pool", singleInstance: true, mergeLive: true}

	t.Run("no flags keep the live scaling", func(t *testing.T) {
		flags := scalingSleepDelayTestFlagSet(sandboxPoolScalingSleepDelayFlagUsage)

		def := liveScalingPoolDefinition(300)
		require.NoError(t, h.parseScalingSleepDelay(flags, def, opts))

		scalings := def.GetScalings()
		require.Len(t, scalings, 1)
		assert.Equal(t, int64(0), scalings[0].GetMin())
		assert.Equal(t, int64(1), scalings[0].GetMax())
		targets := scalings[0].GetTargets()
		require.Len(t, targets, 1)
		delays := targets[0].GetSleepIdleDelay()
		assert.Equal(t, int64(300), delays.GetLightSleepValue())
	})

	t.Run("changed flags override the live scaling", func(t *testing.T) {
		flags := scalingSleepDelayTestFlagSet(sandboxPoolScalingSleepDelayFlagUsage)
		require.NoError(t, flags.Parse([]string{"--min-scale", "2"}))

		def := liveScalingPoolDefinition(0)
		require.NoError(t, h.parseScalingSleepDelay(flags, def, opts))

		scalings := def.GetScalings()
		require.Len(t, scalings, 1)
		assert.Equal(t, int64(2), scalings[0].GetMin())
		assert.Equal(t, int64(1), scalings[0].GetMax(), "the single-instance cap is max-scale 1")
	})

	t.Run("a live sleep delay conflicts with min-scale above 0", func(t *testing.T) {
		flags := scalingSleepDelayTestFlagSet(sandboxPoolScalingSleepDelayFlagUsage)
		require.NoError(t, flags.Parse([]string{"--min-scale", "1"}))

		def := liveScalingPoolDefinition(300)
		err := h.parseScalingSleepDelay(flags, def, opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Error while configuring the pool")
		assert.Contains(t, err.Error(),
			"--light-sleep-delay and --deep-sleep-delay can only be used when min-scale is 0")
	})

	t.Run("a zero sleep delay clears the live delay", func(t *testing.T) {
		flags := scalingSleepDelayTestFlagSet(sandboxPoolScalingSleepDelayFlagUsage)
		require.NoError(t, flags.Parse([]string{"--light-sleep-delay", "0"}))

		def := liveScalingPoolDefinition(300)
		require.NoError(t, h.parseScalingSleepDelay(flags, def, opts))

		scalings := def.GetScalings()
		require.Len(t, scalings, 1)
		assert.Equal(t, int64(0), scalings[0].GetMin())
		assert.Empty(t, scalings[0].GetTargets(), "the cleared delay must drop its target")
	})
}
