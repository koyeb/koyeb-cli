package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/flags_list"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// newPoolCreateCmd builds the `koyeb pool create NAME` command.
func newPoolCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a service pool",
		Long: `Create a service pool.

Pools are project-scoped: use --project (or --workspace) to select the
project the pool is created in. Without it, the request is sent without a
project scope and is effectively required by the server.

A pool pre-provisions instances of the given Docker image so they
can be claimed later with 'koyeb pool claim'.

--type selects the definition the pool members run: "sandbox" (the
default), "web" or "worker". DATABASE pools are not supported. SANDBOX
pools keep the sandbox auto-wiring (the platform owns ports 3030/3031
and mints the executor secret); explicit --port/--route flags are
rejected on them. WEB and WORKER pools carry exactly the declared
--port/--route values, verbatim.`,
		Args: cobra.ExactArgs(1),
		Example: `
# Create a pool of 3 instances
$> koyeb pool create my-pool --size 3 --docker ghcr.io/acme/sandbox

# Create a pool in a specific project
$> koyeb pool create my-pool --project my-project

# Create a WEB pool with explicit member ports and routes
$> koyeb pool create my-pool --type web --port 8080:http --route /:8080
`,
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			req, err := buildCreateServicePool(ctx, cmd, args[0])
			if err != nil {
				return err
			}
			return NewPoolHandler().Create(ctx, cmd, args, req)
		}),
	}

	addPoolFlags(cmd.Flags())

	return cmd
}

// buildCreateServicePool builds the create request. parseServiceDefinitionFlags
// is never called: it dereferences flags this command does not register.
func buildCreateServicePool(ctx *CLIContext, cmd *cobra.Command, name string) (koyeb.CreateServicePool, error) {
	flags := cmd.Flags()
	svcHandler := NewServiceHandler()

	poolType, err := parsePoolType(flags)
	if err != nil {
		return koyeb.CreateServicePool{}, err
	}
	if err := validatePoolWiringFlags(poolType, flags); err != nil {
		return koyeb.CreateServicePool{}, err
	}
	if err := validatePoolSandboxKnobs(poolType, flags); err != nil {
		return koyeb.CreateServicePool{}, err
	}

	def := koyeb.NewDeploymentDefinitionWithDefaults()

	// Docker source: the shared bundle parses the flags and defaults the
	// image to koyeb/sandbox when --docker is unset. The builder stays
	// pure: no image verification API call on the pool paths.
	dockerSource := koyeb.NewDockerSourceWithDefaults()
	parsedDocker, _, err := svcHandler.parseDockerSource(ctx, flags, dockerSource, dockerSourceParseOptions{
		defaultImage: koyebSandboxImage,
	})
	if err != nil {
		return koyeb.CreateServicePool{}, err
	}
	def.SetDocker(*parsedDocker)

	// Instance type and regions: the shared bundle applies both.
	if err := svcHandler.parseInstanceTypeRegions(flags, def); err != nil {
		return koyeb.CreateServicePool{}, err
	}

	// Environment variables and config files: the shared bundle applies
	// both.
	if err := svcHandler.parseEnvConfigFiles(ctx, flags, def); err != nil {
		return koyeb.CreateServicePool{}, err
	}

	// Member network policy (egress).
	networkPolicy, policyChanged, err := svcHandler.parseNetworkPolicy(flags, nil)
	if err != nil {
		return koyeb.CreateServicePool{}, err
	}
	if policyChanged && networkPolicy != nil {
		def.SetNetworkPolicy(*networkPolicy)
	}

	// Pools run at max-scale=1 with the same curated flag set as
	// sandboxes: the shared bundle applies the single-instance scaling.
	if err := svcHandler.parseScalingSleepDelay(flags, def, scalingSleepDelayParseOptions{
		what:           "pool",
		singleInstance: true,
	}); err != nil {
		return koyeb.CreateServicePool{}, err
	}

	def.SetType(poolType)

	// SANDBOX wiring (ports 3030/3031 and the sandbox routes) stays
	// server-owned; non-SANDBOX pools carry the declared values verbatim.
	if err := setPoolPortsAndRoutes(poolType, flags, def); err != nil {
		return koyeb.CreateServicePool{}, err
	}
	if err := applyPoolSandboxKnobs(poolType, flags, def); err != nil {
		return koyeb.CreateServicePool{}, err
	}

	// The server requires the definition name to be set (SANDBOX case).
	def.SetName(name)

	size, _ := flags.GetInt64("size")

	return koyeb.CreateServicePool{
		Name:       &name,
		Size:       &size,
		Definition: def,
	}, nil
}

// applyPoolSandboxKnobs applies the sandbox-only knobs on SANDBOX pools:
// the platform owns the default wiring, so an explicit protocol request
// emits the wiring with the chosen protocol (python/JS parity); the TCP
// proxy is a standalone proxy-ports entry and needs no wiring.
func applyPoolSandboxKnobs(poolType koyeb.DeploymentDefinitionType, flags *pflag.FlagSet,
	def *koyeb.DeploymentDefinition) error {
	if poolType != koyeb.DEPLOYMENTDEFINITIONTYPE_SANDBOX {
		return nil
	}
	if flags.Lookup("exposed-port-protocol").Changed {
		protocol, _ := flags.GetString("exposed-port-protocol")
		if protocol != "http" && protocol != "http2" {
			return &errors.CLIError{
				What:     "Invalid exposed port protocol",
				Why:      fmt.Sprintf("Invalid protocol '%s'. Must be one of ('http', 'http2')", protocol),
				Orig:     nil,
				Solution: "Use --exposed-port-protocol http or --exposed-port-protocol http2",
			}
		}
		configureSandboxPortsAndRoutes(def, protocol)
	}
	if flags.Lookup("enable-tcp-proxy").Changed {
		enableTCPProxy, _ := flags.GetBool("enable-tcp-proxy")
		def.SetProxyPorts(mergeTCPProxyPort(def.GetProxyPorts(), enableTCPProxy))
	}
	return nil
}

// mergeTCPProxyPort merges the --enable-tcp-proxy state into the live
// proxy ports: enabling adds the executor's 3031/tcp entry, disabling
// removes it. Other proxy ports the pool may carry are untouched.
func mergeTCPProxyPort(current []koyeb.DeploymentProxyPort, enable bool) []koyeb.DeploymentProxyPort {
	merged := make([]koyeb.DeploymentProxyPort, 0, len(current)+1)
	for _, proxyPort := range current {
		if proxyPort.GetPort() != 3031 {
			merged = append(merged, proxyPort)
		}
	}
	if enable {
		port := int64(3031)
		protocol := koyeb.PROXYPORTPROTOCOL_TCP
		merged = append(merged, koyeb.DeploymentProxyPort{Port: &port, Protocol: &protocol})
	}
	return merged
}

// setPoolPortsAndRoutes parses the --port and --route flags onto the
// definition. WEB and WORKER pools carry the declared values verbatim;
// SANDBOX wiring (ports 3030/3031 and the sandbox routes) stays
// server-owned, so SANDBOX pool definitions are never touched here —
// validatePoolWiringFlags rejects the flags first.
func setPoolPortsAndRoutes(poolType koyeb.DeploymentDefinitionType,
	flags *pflag.FlagSet, def *koyeb.DeploymentDefinition) error {
	if poolType == koyeb.DEPLOYMENTDEFINITIONTYPE_SANDBOX {
		return nil
	}

	ports, err := parseListFlags("ports", flags_list.NewPortListFromFlags, flags, def.Ports)
	if err != nil {
		return err
	}
	// Set on change even when the merge empties the list: deleting the
	// last port must clear the wiring, not silently no-op (the len>0
	// guard would swallow it).
	if flags.Lookup("ports").Changed || len(ports) > 0 {
		def.SetPorts(ports)
	}

	routes, err := parseListFlags("routes", flags_list.NewRouteListFromFlags, flags, def.Routes)
	if err != nil {
		return err
	}
	if flags.Lookup("routes").Changed || len(routes) > 0 {
		def.SetRoutes(routes)
	}
	return nil
}

// Create creates a service pool.
func (h *PoolHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string, req koyeb.CreateServicePool) error {
	res, resp, err := ctx.API.CreateServicePool(ctx.Context, req)
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while creating the pool `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	poolReply := NewPoolReply(ctx.Mapper, res.GetServicePool(), full)
	ctx.Renderer.Render(poolReply)
	return nil
}
