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

	def := koyeb.NewDeploymentDefinitionWithDefaults()

	// Docker source
	dockerSource := koyeb.NewDockerSourceWithDefaults()
	if flags.Lookup("docker-private-registry-secret").Changed {
		secret, _ := flags.GetString("docker-private-registry-secret")
		dockerSource.SetImageRegistrySecret(secret)
	}
	if flags.Lookup("docker").Changed {
		image, _ := flags.GetString("docker")
		dockerSource.SetImage(image)
	} else {
		dockerSource.SetImage("koyeb/sandbox")
	}
	if flags.Lookup("docker-args").Changed {
		args, _ := flags.GetStringSlice("docker-args")
		dockerSource.SetArgs(args)
	}
	if flags.Lookup("docker-command").Changed {
		command, _ := flags.GetString("docker-command")
		dockerSource.SetCommand(command)
	}
	if flags.Lookup("docker-entrypoint").Changed {
		entrypoint, _ := flags.GetStringSlice("docker-entrypoint")
		dockerSource.SetEntrypoint(entrypoint)
	}
	def.SetDocker(*dockerSource)

	// Instance type
	def.SetInstanceTypes(svcHandler.parseInstanceType(flags, nil))

	// Regions
	regions, err := svcHandler.parseRegions(flags, nil)
	if err != nil {
		return koyeb.CreateServicePool{}, err
	}
	def.SetRegions(regions)

	// Environment variables
	envVars, err := svcHandler.parseEnv(flags, nil)
	if err != nil {
		return koyeb.CreateServicePool{}, err
	}
	def.SetEnv(envVars)

	// Config files
	parsedFiles, err := svcHandler.parseConfigFiles(ctx, flags, nil)
	if err != nil {
		return koyeb.CreateServicePool{}, err
	}
	def.SetConfigFiles(parsedFiles)

	// Pools run at max-scale=1 with the same curated flag set as sandboxes.
	scaling, err := parseSingleInstanceScaling(flags, "pool")
	if err != nil {
		return koyeb.CreateServicePool{}, err
	}
	def.SetScalings([]koyeb.DeploymentScaling{scaling})

	def.SetType(poolType)

	// SANDBOX wiring (ports 3030/3031 and the sandbox routes) stays
	// server-owned; non-SANDBOX pools carry the declared values verbatim.
	if err := setPoolPortsAndRoutes(poolType, flags, def); err != nil {
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
	if len(ports) > 0 {
		def.SetPorts(ports)
	}

	routes, err := parseListFlags("routes", flags_list.NewRouteListFromFlags, flags, def.Routes)
	if err != nil {
		return err
	}
	if len(routes) > 0 {
		def.SetRoutes(routes)
	}
	return nil
}

// Create creates a service pool.
func (h *PoolHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string, req koyeb.CreateServicePool) error {
	res, resp, err := ctx.Client.ServicePoolsApi.CreateServicePool(ctx.Context).ServicePool(req).Execute()
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
