package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// newPoolUpdateCmd builds the `koyeb pool update NAME` command.
func newPoolUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update a service pool",
		Long: `Update a service pool.

The update is a full replace: the live pool is refetched, the changed
flags are applied over its definition, and size + definition are resent
in one PUT (the API rejects update_mask and requires the definition).
Flags that are not passed keep their current values.

The NAME argument identifies the pool (name, short ID or full UUID);
renaming a pool is not supported. The pool type is fixed at creation:
--type must match the live type (restating it is allowed).`,
		Args: cobra.ExactArgs(1),
		Example: `
# Resize a pool
$> koyeb pool update my-pool --size 5

# Update the image of a pool's members
$> koyeb pool update my-pool --docker ghcr.io/acme/sandbox:v2

# Upsert an environment variable on a WEB pool's members
$> koyeb pool update my-pool --env LOG_LEVEL=debug
`,
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			poolID, err := ResolvePoolArgs(ctx, args[0])
			if err != nil {
				return err
			}
			return NewPoolHandler().Update(ctx, cmd, args, poolID)
		}),
	}

	addPoolFlags(cmd.Flags())

	return cmd
}

// Update reconfigures a service pool with a full-replace PUT: the API
// rejects update_mask and requires the definition on every update, so the
// live pool is refetched and size + definition resent with the changed
// flags applied — mirroring the python SDK's ServicePool.update.
func (h *PoolHandler) Update(ctx *CLIContext, cmd *cobra.Command, args []string, poolID string) error {
	current, resp, err := ctx.API.GetServicePool(ctx.Context, poolID)
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the pool `%s`", args[0]),
			err,
			resp,
		)
	}

	req, err := buildUpdateServicePool(ctx, cmd.Flags(), current.GetServicePool(), args[0])
	if err != nil {
		return err
	}

	res, resp, err := ctx.API.UpdateServicePool(ctx.Context, poolID, req)
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while updating the pool `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	poolReply := NewPoolReply(ctx.Mapper, res.GetServicePool(), full)
	ctx.Renderer.Render(poolReply)
	return nil
}

// effectivePoolType returns the pool type the update applies: the --type
// flag when set, otherwise the live definition's type.
func effectivePoolType(flags *pflag.FlagSet, current koyeb.DeploymentDefinitionType) (
	koyeb.DeploymentDefinitionType, error) {
	if !flags.Lookup("type").Changed {
		return current, nil
	}
	return parsePoolType(flags)
}

// buildUpdateServicePool builds the full-replace update request from the
// live pool: changed flags are applied over the current definition, then
// size and definition are resent together.
func buildUpdateServicePool(ctx *CLIContext, flags *pflag.FlagSet, current koyeb.ServicePool, name string) (
	koyeb.UpdateServicePool, error) {
	if !current.HasDefinition() {
		return koyeb.UpdateServicePool{}, &errors.CLIError{
			What: "Error while updating the pool",
			Why:  fmt.Sprintf("the pool `%s` has no definition to update", name),
			Additional: []string{
				"The update endpoint is a full replace: it requires the pool's current definition.",
			},
			Orig:     nil,
			Solution: errors.CLIErrorSolution("Fetch the pool with `koyeb pool describe " + name + "` and try again"),
		}
	}
	def := current.GetDefinition()

	svcHandler := NewServiceHandler()

	liveType := def.GetType()
	poolType, err := effectivePoolType(flags, liveType)
	if err != nil {
		return koyeb.UpdateServicePool{}, err
	}
	// The type is fixed at creation: a full replace that re-types the live
	// definition would leave the members mis-wired (a WEB pool still
	// carrying the sandbox wiring and its executor secret).
	if poolType != liveType {
		return koyeb.UpdateServicePool{}, &errors.CLIError{
			What: "Error while updating the pool",
			Why:  "the pool type cannot be changed on update",
			Additional: []string{
				"The update is a full replace of the live definition: re-typing it in place would leave the members mis-wired.",
				"Recreate the pool to change its type: `koyeb pool delete`, then `koyeb pool create --type <type>`.",
			},
			Orig:     nil,
			Solution: "Remove the --type flag or set it to the pool's current type, and try again",
		}
	}
	if err := validatePoolSourceFlags(flags); err != nil {
		return koyeb.UpdateServicePool{}, err
	}
	if err := validatePoolWiringFlags(poolType, flags); err != nil {
		return koyeb.UpdateServicePool{}, err
	}
	if err := validatePoolSandboxKnobs(poolType, flags); err != nil {
		return koyeb.UpdateServicePool{}, err
	}
	def.SetType(poolType)

	// Member source: the docker and archive bundles are mutually
	// exclusive (validatePoolSourceFlags). A changed bundle replaces the
	// live source; without source flags the shared --privileged applies
	// to the live source's builders. The pure options keep the builders
	// free of API calls.
	switch {
	case hasPoolArchiveSourceFlags(flags):
		archive := def.GetArchive()
		parsedArchive, err := svcHandler.parseArchiveSource(flags, &archive)
		if err != nil {
			return koyeb.UpdateServicePool{}, err
		}
		def.SetArchive(*parsedArchive)
		def.Docker = nil
	case hasPoolDockerSourceFlags(flags):
		dockerSource := def.GetDocker()
		parsedDocker, _, err := svcHandler.parseDockerSource(ctx, flags, &dockerSource,
			dockerSourceParseOptions{})
		if err != nil {
			return koyeb.UpdateServicePool{}, err
		}
		def.SetDocker(*parsedDocker)
		def.Archive = nil
	case def.HasArchive():
		// No source flag passed: keep the live archive and apply the
		// shared flags to its builder.
		archive := def.GetArchive()
		parsedArchive, err := svcHandler.parseArchiveSource(flags, &archive)
		if err != nil {
			return koyeb.UpdateServicePool{}, err
		}
		def.SetArchive(*parsedArchive)
	default:
		dockerSource := def.GetDocker()
		parsedDocker, dockerChanged, err := svcHandler.parseDockerSource(ctx, flags, &dockerSource,
			dockerSourceParseOptions{})
		if err != nil {
			return koyeb.UpdateServicePool{}, err
		}
		if dockerChanged {
			def.SetDocker(*parsedDocker)
		}
	}

	// Instance type and regions: the shared bundle merges both over the
	// live values.
	if err := svcHandler.parseInstanceTypeRegions(flags, &def); err != nil {
		return koyeb.UpdateServicePool{}, err
	}

	if err := svcHandler.parseEnvConfigFiles(ctx, flags, &def); err != nil {
		return koyeb.UpdateServicePool{}, err
	}

	// Member volumes: the shared bundle merges the flags over the live
	// mounts (unchanged flags keep the live values).
	volumes, err := svcHandler.parseVolumes(ctx, flags, def.Volumes)
	if err != nil {
		return koyeb.UpdateServicePool{}, err
	}
	def.SetVolumes(volumes)

	// Scaling and sleep delays: the shared bundle merges over the live
	// scaling (unchanged flags keep the live values).
	if err := svcHandler.parseScalingSleepDelay(flags, &def, scalingSleepDelayParseOptions{
		what:           "pool",
		singleInstance: true,
		mergeLive:      true,
	}); err != nil {
		return koyeb.UpdateServicePool{}, err
	}

	// Member network policy (egress), merged over the live policy.
	var currentPolicy *koyeb.NetworkPolicy
	if def.HasNetworkPolicy() {
		np := def.GetNetworkPolicy()
		currentPolicy = &np
	}
	networkPolicy, policyChanged, err := svcHandler.parseNetworkPolicy(flags, currentPolicy)
	if err != nil {
		return koyeb.UpdateServicePool{}, err
	}
	if policyChanged && networkPolicy != nil {
		def.SetNetworkPolicy(*networkPolicy)
	}

	// SANDBOX wiring (ports 3030/3031 and the sandbox routes) stays
	// server-owned; non-SANDBOX pools merge the declared values verbatim.
	if err := setPoolPortsAndRoutes(poolType, flags, &def); err != nil {
		return koyeb.UpdateServicePool{}, err
	}
	if err := setPoolProxyPorts(poolType, flags, &def); err != nil {
		return koyeb.UpdateServicePool{}, err
	}
	if err := applyPoolSandboxKnobs(poolType, flags, &def); err != nil {
		return koyeb.UpdateServicePool{}, err
	}

	// Member healthchecks, merged over the live checks (unchanged flags
	// keep the live values): the shared bundle parses --checks and
	// --checks-grace-period; parseChecks rejects them on non-WEB pools,
	// matching the service surfaces.
	healthChecks, err := svcHandler.parseChecks(poolType, flags, def.HealthChecks)
	if err != nil {
		return koyeb.UpdateServicePool{}, err
	}
	def.SetHealthChecks(healthChecks)

	size := current.GetSize()
	if flags.Lookup("size").Changed {
		size, _ = flags.GetInt64("size")
	}

	return koyeb.UpdateServicePool{
		Size:       &size,
		Definition: &def,
	}, nil
}

// mergePoolScalings merges the --min-scale and sleep-delay flags into the
// pool's current scaling. Pools always run at max-scale 1. Unlike create,
// unchanged flags keep the live scaling untouched; the live sleep delays
// carry over and changed flags override or clear them.
func mergePoolScalings(flags *pflag.FlagSet, current []koyeb.DeploymentScaling) (
	[]koyeb.DeploymentScaling, error) {
	scalingChanged := flags.Lookup("min-scale").Changed ||
		flags.Lookup("light-sleep-delay").Changed ||
		flags.Lookup("deep-sleep-delay").Changed
	if !scalingChanged {
		return current, nil
	}

	minScale := int64(1)
	light, deep := int64(0), int64(0)
	if len(current) > 0 {
		minScale = current[0].GetMin()
		light, deep = currentSleepDelays(current[0].GetTargets())
	}
	if flags.Lookup("min-scale").Changed {
		minScale, _ = flags.GetInt64("min-scale")
	}
	if flags.Lookup("light-sleep-delay").Changed {
		light = sleepDelayFlagSeconds(flags, "light-sleep-delay")
	}
	if flags.Lookup("deep-sleep-delay").Changed {
		deep = sleepDelayFlagSeconds(flags, "deep-sleep-delay")
	}

	if (light > 0 || deep > 0) && minScale > 0 {
		return nil, &errors.CLIError{
			What: "Error while configuring the pool",
			Why:  "--light-sleep-delay and --deep-sleep-delay can only be used when min-scale is 0",
			Additional: []string{
				"Sleep delays are only applicable to services that can scale to zero.",
				"Clear the sleep delays with --light-sleep-delay 0 and/or --deep-sleep-delay 0, or keep --min-scale 0.",
			},
			Orig:     nil,
			Solution: "Set --min-scale 0 or clear the sleep delays, and try again",
		}
	}

	scaling := koyeb.NewDeploymentScalingWithDefaults()
	scaling.SetMin(minScale)
	scaling.SetMax(1)
	if light > 0 || deep > 0 {
		delays := koyeb.NewDeploymentScalingTargetSleepIdleDelay()
		if light > 0 {
			delays.SetLightSleepValue(light)
		}
		if deep > 0 {
			delays.SetDeepSleepValue(deep)
		}
		target := koyeb.NewDeploymentScalingTarget()
		target.SetSleepIdleDelay(*delays)
		scaling.Targets = []koyeb.DeploymentScalingTarget{*target}
	}

	return []koyeb.DeploymentScaling{*scaling}, nil
}

// currentSleepDelays extracts the configured light and deep sleep delays
// (in seconds) from the scaling targets.
func currentSleepDelays(targets []koyeb.DeploymentScalingTarget) (light, deep int64) {
	for i := range targets {
		delays := targets[i].GetSleepIdleDelay()
		if delays.HasLightSleepValue() {
			light = delays.GetLightSleepValue()
		}
		if delays.HasDeepSleepValue() {
			deep = delays.GetDeepSleepValue()
		}
	}
	return light, deep
}

// sleepDelayFlagSeconds converts a sleep-delay duration flag to seconds;
// a zero duration clears the delay.
func sleepDelayFlagSeconds(flags *pflag.FlagSet, name string) int64 {
	duration, _ := flags.GetDuration(name)
	if duration > 0 {
		return int64(duration.Seconds())
	}
	return 0
}
