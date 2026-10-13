package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// sandboxUpdateErrorWhat prefixes the `sandbox update` errors.
const sandboxUpdateErrorWhat = "Error while updating the sandbox"

// Update updates a running sandbox's lifecycle and egress network policy,
// mirroring the SDKs' update_lifecycle and update_network_policy.
func (h *SandboxHandler) Update(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	return updateSandbox(ctx, cmd, args)
}

// updateSandbox runs the update flow: refuse a flag-less update, resolve
// the sandbox, apply the lifecycle and network policy flags over the
// latest deployment, and send the update — which redeploys the sandbox.
func updateSandbox(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	// UpdateService redeploys the sandbox even with no change, so a
	// flag-less update is refused up front, before any API round-trip.
	if !sandboxUpdateFlagsChanged(cmd.Flags()) {
		return &errors.CLIError{
			What: sandboxUpdateErrorWhat,
			Why:  "nothing to update: no update flag was provided",
			Additional: []string{
				"Lifecycle: --delete-after-delay, --delete-after-inactivity-delay.",
				"Network policy: --block-network, --outbound-allowlist, --no-network-policy.",
				"An update always redeploys the sandbox, so it must change something.",
			},
			Orig:     nil,
			Solution: "Pass at least one update flag and try again",
		}
	}

	service, err := resolveSandboxService(ctx, args[0])
	if err != nil {
		return err
	}

	svcHandler := NewServiceHandler()
	updateService, policyChanged, err := buildSandboxUpdateService(ctx, cmd, service, args[0], svcHandler)
	if err != nil {
		return err
	}

	if policyChanged {
		// SDK parity: update_network_policy warns that the sandbox is
		// restarted and in-memory state is lost.
		log.Warnf(
			"Applying a new network policy redeploys the sandbox: it is restarted and any in-memory or non-persisted state is lost",
		)
	}

	if _, resp, err := ctx.API.UpdateService(ctx.Context, service.GetId(), *updateService); err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("%s `%s`", sandboxUpdateErrorWhat, args[0]),
			err,
			resp,
		)
	}

	log.Infof(
		"Sandbox update in progress. To access the build logs, run: `koyeb service logs %s -t build`. For the runtime logs, run `koyeb service logs %s`",
		service.GetId()[:8],
		service.GetId()[:8],
	)
	renderServiceState(ctx, cmd, service.GetId())
	return nil
}

// buildSandboxUpdateService builds the update request the way the SDKs do:
// the latest deployment's definition with the network policy flags applied
// to it, and the service's lifecycle merged with the lifecycle flags —
// only the supplied fields change (merge-not-clobber). Returns whether the
// egress network policy changed.
func buildSandboxUpdateService(
	ctx *CLIContext,
	cmd *cobra.Command,
	service *koyeb.Service,
	name string,
	svcHandler *ServiceHandler,
) (*koyeb.UpdateService, bool, error) {
	deployment, err := latestSandboxDeployment(ctx, service, name)
	if err != nil {
		return nil, false, err
	}
	definition := deployment.GetDefinition()

	// Member network policy (egress), merged over the live policy like on
	// `pool update`: unchanged flags keep the live policy.
	var currentPolicy *koyeb.NetworkPolicy
	if definition.HasNetworkPolicy() {
		policy := definition.GetNetworkPolicy()
		currentPolicy = &policy
	}
	networkPolicy, policyChanged, err := svcHandler.parseNetworkPolicy(cmd.Flags(), currentPolicy)
	if err != nil {
		return nil, false, err
	}
	if policyChanged && networkPolicy != nil {
		definition.SetNetworkPolicy(*networkPolicy)
	}

	updateService := koyeb.NewUpdateServiceWithDefaults()
	updateService.SetDefinition(definition)

	// Lifecycle merge: the supplied flags update their fields, the others
	// keep the live values; the live lifecycle is resent when it exists
	// (Python parity: service.life_cycle is always sent back).
	var currentLifeCycle *koyeb.ServiceLifeCycle
	if service.HasLifeCycle() {
		lifecycle := service.GetLifeCycle()
		currentLifeCycle = &lifecycle
	}
	if lifecycle := svcHandler.parseLifeCycle(cmd.Flags(), currentLifeCycle); lifecycle != nil {
		updateService.SetLifeCycle(*lifecycle)
	}

	return updateService, policyChanged, nil
}

// latestSandboxDeployment fetches the deployment the update is based on,
// mirroring the SDKs' get_deployment(service.latest_deployment_id).
func latestSandboxDeployment(ctx *CLIContext, service *koyeb.Service, name string) (*koyeb.Deployment, error) {
	res, resp, err := ctx.API.GetDeployment(ctx.Context, service.GetLatestDeploymentId())
	if err != nil {
		return nil, errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while updating the sandbox `%s`", name),
			err,
			resp,
		)
	}

	deployment := res.GetDeployment()
	if !deployment.HasDefinition() {
		return nil, &errors.CLIError{
			What: sandboxUpdateErrorWhat,
			Why:  fmt.Sprintf("we couldn't find the latest deployment of the sandbox `%s`", name),
			Additional: []string{
				"The update is based on the latest deployment's definition.",
			},
			Orig:     nil,
			Solution: "Wait for the sandbox to be deployed and try again",
		}
	}
	return &deployment, nil
}

// sandboxUpdateFlagNames lists the flags `sandbox update` accepts.
var sandboxUpdateFlagNames = []string{
	"delete-after-delay", "delete-after-inactivity-delay",
	"block-network", "outbound-allowlist", "no-network-policy",
}

// sandboxUpdateFlagsChanged reports whether any sandbox update flag was
// provided.
func sandboxUpdateFlagsChanged(flags *pflag.FlagSet) bool {
	for _, name := range sandboxUpdateFlagNames {
		if flags.Lookup(name).Changed {
			return true
		}
	}
	return false
}
