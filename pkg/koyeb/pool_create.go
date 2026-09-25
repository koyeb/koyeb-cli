package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
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
can be claimed later with 'koyeb pool claim'.`,
		Args: cobra.ExactArgs(1),
		Example: `
# Create a pool of 3 instances
$> koyeb pool create my-pool --size 3 --docker ghcr.io/acme/sandbox

# Create a pool in a specific project
$> koyeb pool create my-pool --project my-project
`,
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			if err := setProjectHeader(ctx, cmd); err != nil {
				return err
			}
			req, err := buildCreateServicePool(ctx, cmd, args[0])
			if err != nil {
				return err
			}
			return NewPoolHandler().Create(ctx, cmd, args, req)
		}),
	}

	addPoolCreateFlags(cmd.Flags())

	return cmd
}

func addPoolCreateFlags(flags *pflag.FlagSet) {
	flags.Int64("size", 1, "Number of instances kept ready in the pool")

	flags.String("docker", "", "Docker image (default: koyeb/sandbox)")
	flags.String("docker-private-registry-secret", "", "Docker private registry secret")
	flags.StringSlice("docker-entrypoint", []string{}, "Docker entrypoint")
	flags.String("docker-command", "", "Docker command")
	flags.StringSlice("docker-args", []string{}, "Docker command arguments")

	flags.String("instance-type", "micro", "Instance type")
	flags.StringSlice("regions", []string{}, "Deployment regions")

	flags.StringSlice("env", []string{}, "Environment variables (KEY=VALUE)")
	flags.StringSlice("config-file", nil, "Config files (LOCAL:REMOTE:PERMS)")

	flags.Int64("min-scale", 1, "Min scale")

	flags.Duration("light-sleep-delay", 0,
		"Delay after which an idle service is put to light sleep. "+
			"Use duration format (e.g., '1m', '5m', '1h'). Set to 0 to disable.")
	flags.Duration("deep-sleep-delay", 0,
		"Delay after which an idle service is put to deep sleep. "+
			"Use duration format (e.g., '5m', '30m', '1h'). Set to 0 to disable.")
}

// buildCreateServicePool builds the create request. parseServiceDefinitionFlags
// is never called: it dereferences flags this command does not register.
func buildCreateServicePool(ctx *CLIContext, cmd *cobra.Command, name string) (koyeb.CreateServicePool, error) {
	flags := cmd.Flags()
	svcHandler := NewServiceHandler()

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

	def.SetType(koyeb.DEPLOYMENTDEFINITIONTYPE_SANDBOX)
	// The server requires the definition name to be set (SANDBOX case).
	def.SetName(name)

	size, _ := flags.GetInt64("size")

	return koyeb.CreateServicePool{
		Name:       &name,
		Size:       &size,
		Definition: def,
	}, nil
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
