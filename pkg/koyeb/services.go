package koyeb

import (
	"fmt"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/flags_list"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const ServiceExamples = `
# Deploy a nginx docker image, with default port (80:http), default route (/:80)
$> koyeb service create myservice --app myapp --docker nginx

# Build and deploy a GitHub repository using buildpack (default), set the environment variable PORT, and expose the port 8000 to the root route
$> koyeb service create myservice --app myapp --git github.com/koyeb/example-flask --git-branch main --env PORT=8000 --port 8000:http --route /:8000

# Build and deploy a GitHub repository using docker
$> koyeb service create myservice --app myapp --git github.com/org/name --git-branch main --git-builder docker
`

func NewServiceCmd() *cobra.Command {
	h := NewServiceHandler()

	serviceCmd := &cobra.Command{
		Use:     "services ACTION",
		Aliases: []string{"s", "svc", "service"},
		Short:   "Services",
	}

	createServiceCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create service",
		Args:  cobra.ExactArgs(1),
		Example: `
# Deploy a nginx docker image, with default port (80:http), default route (/:80)
$> koyeb service create myservice --app myapp --docker nginx

# Build and deploy a GitHub repository using buildpack (default), set the environment variable PORT, and expose the port 8000 to the root route
$> koyeb service create myservice --app myapp --git github.com/koyeb/example-flask --git-branch main --env PORT=8000 --port 8000:http --route /:8000

# Build and deploy a GitHub repository using docker
$> koyeb service create myservice --app myapp --git github.com/org/name --git-branch main --git-builder docker
`,
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			createService := koyeb.NewCreateServiceWithDefaults()
			createDefinition := koyeb.NewDeploymentDefinitionWithDefaults()

			err := parseServiceDefinitionFlags(cmd.Flags(), createDefinition, true)
			if err != nil {
				return err
			}
			createDefinition.Name = koyeb.PtrString(args[0])

			createService.SetDefinition(*createDefinition)
			return h.Create(ctx, cmd, args, createService)
		}),
	}
	addServiceDefinitionFlags(createServiceCmd.Flags())
	createServiceCmd.Flags().StringP("app", "a", "", "Service application")
	createServiceCmd.MarkFlagRequired("app") //nolint:errcheck
	serviceCmd.AddCommand(createServiceCmd)

	getServiceCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get service",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Get),
	}
	serviceCmd.AddCommand(getServiceCmd)

	logsServiceCmd := &cobra.Command{
		Use:     "logs NAME",
		Aliases: []string{"l", "log"},
		Short:   "Get the service logs",
		Args:    cobra.ExactArgs(1),
		RunE:    WithCLIContext(h.Logs),
	}
	serviceCmd.AddCommand(logsServiceCmd)
	logsServiceCmd.Flags().String("instance", "", "Instance")
	logsServiceCmd.Flags().StringP("type", "t", "", "Type (runtime,build)")

	listServiceCmd := &cobra.Command{
		Use:   "list",
		Short: "List services",
		RunE:  WithCLIContext(h.List),
	}
	serviceCmd.AddCommand(listServiceCmd)
	listServiceCmd.Flags().StringP("app", "a", "", "App")
	listServiceCmd.Flags().StringP("name", "n", "", "Service name")

	describeServiceCmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe service",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Describe),
	}
	serviceCmd.AddCommand(describeServiceCmd)

	execServiceCmd := &cobra.Command{
		Use:     "exec NAME CMD -- [args...]",
		Short:   "Run a command in the context of an instance selected among the service instances",
		Aliases: []string{"run", "attach"},
		Args:    cobra.MinimumNArgs(2),
		RunE:    WithCLIContext(h.Exec),
	}
	serviceCmd.AddCommand(execServiceCmd)

	updateServiceCmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update service",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			service, err := h.ResolveServiceArgs(ctx, args[0])
			if err != nil {
				return err
			}

			updateService := koyeb.NewUpdateServiceWithDefaults()
			latestDeploy, resp, err := ctx.Client.DeploymentsApi.
				ListDeployments(ctx.Context).
				Limit("1").
				ServiceId(service).
				Execute()

			if err != nil {
				return errors.NewCLIErrorFromAPIError(
					fmt.Sprintf("Error while updating the service `%s`", args[0]),
					err,
					resp,
				)
			}
			if len(latestDeploy.GetDeployments()) == 0 {
				return &errors.CLIError{
					What: "Error while updating the service",
					Why:  "we couldn't find the latest deployment of your service",
					Additional: []string{
						"When you create a service for the first time, it can take a few seconds for the first deployment to be created.",
						"We need to fetch the configuration of this latest deployment to update your service.",
					},
					Orig:     nil,
					Solution: "Try again in a few seconds. If the problem persists, please create an issue on https://github.com/koyeb/koyeb-cli/issues/new",
				}
			}
			updateDef := latestDeploy.GetDeployments()[0].Definition
			err = parseServiceDefinitionFlags(cmd.Flags(), updateDef, false)
			if err != nil {
				return err
			}
			updateService.SetDefinition(*updateDef)
			return h.Update(ctx, cmd, args, updateService)
		}),
	}
	addServiceDefinitionFlags(updateServiceCmd.Flags())
	serviceCmd.AddCommand(updateServiceCmd)

	redeployServiceCmd := &cobra.Command{
		Use:   "redeploy NAME",
		Short: "Redeploy service",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.ReDeploy),
	}
	serviceCmd.AddCommand(redeployServiceCmd)
	redeployServiceCmd.Flags().Bool("use-cache", false, "Use cache to redeploy")

	deleteServiceCmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete service",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Delete),
	}
	serviceCmd.AddCommand(deleteServiceCmd)

	pauseServiceCmd := &cobra.Command{
		Use:   "pause NAME",
		Short: "Pause service",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Pause),
	}
	serviceCmd.AddCommand(pauseServiceCmd)

	resumeServiceCmd := &cobra.Command{
		Use:   "resume NAME",
		Short: "Resume service",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Resume),
	}
	serviceCmd.AddCommand(resumeServiceCmd)

	return serviceCmd
}

func NewServiceHandler() *ServiceHandler {
	return &ServiceHandler{}
}

type ServiceHandler struct {
}

func (h *ServiceHandler) ResolveServiceArgs(ctx *CLIContext, val string) (string, error) {
	serviceMapper := ctx.Mapper.Service()
	id, err := serviceMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (h *ServiceHandler) ResolveAppArgs(ctx *CLIContext, val string) (string, error) {
	appMapper := ctx.Mapper.App()
	id, err := appMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}

func addServiceDefinitionFlags(flags *pflag.FlagSet) {
	flags.String("type", "WEB", `Service type, either "web" or "worker"`)
	flags.String("git", "", "Git repository")
	flags.String("git-branch", "", "Git branch")
	flags.String("git-build-command", "", "Buid command (legacy, prefer git-buildpack-build-command)")
	flags.String("git-run-command", "", "Run command (legacy, prefer git-buildpack-run-command)")
	flags.Bool("git-no-deploy-on-push", false, "Disable new deployments creation when code changes are pushed on the configured branch")
	flags.String("git-workdir", "", "Path to the sub-directory containing the code to build and deploy")

	flags.String("git-builder", "buildpack", `Builder to use, either "buildpack" (default) or "docker"`)

	flags.String("git-buildpack-build-command", "", "Buid command")
	flags.String("git-buildpack-run-command", "", "Run command")

	flags.String("git-docker-dockerfile", "", "Dockerfile path")
	flags.StringSlice("git-docker-entrypoint", []string{}, "Docker entrypoint")
	flags.String("git-docker-command", "", "Docker CMD")
	flags.StringSlice("git-docker-args", []string{}, "Arguments for the Docker CMD")
	flags.String("git-docker-target", "", "Docker target")

	flags.String("docker", "", "Docker image")
	flags.String("docker-private-registry-secret", "", "Docker private registry secret")
	flags.StringSlice("docker-entrypoint", []string{}, "Docker entrypoint")
	flags.String("docker-command", "", "Docker command")
	flags.StringSlice("docker-args", []string{}, "Docker args")
	flags.StringSlice("regions", []string{"fra"}, "Regions")
	flags.StringSlice(
		"env",
		[]string{},
		"Update service environment variables using the format KEY=VALUE, for example --env FOO=bar\n"+
			"To use the value of a secret as an environment variable, specify the secret name preceded by @, for example --env FOO=@bar\n"+
			"To delete an environment variable, prefix its name with '!', for example --env '!FOO'",
	)
	flags.StringSlice(
		"routes",
		[]string{"/:80"},
		"Update service routes (available for services of type \"web\" only) using the format PATH[:PORT], for example '/foo:8080'\n"+
			"If no port is specified, it defaults to 80\n"+
			"To delete a route, use '!PATH', for example --route '!/foo'\n",
	)
	flags.StringSlice(
		"ports",
		[]string{"80:http"},
		"Update service ports (available for services of type \"web\" only) using the format PORT[:PROTOCOL], for example --port 80:http\n"+
			"If no protocol is specified, it defaults to \"http\". Supported protocols are \"http\" and \"http2\"\n"+
			"To delete an exposed port, prefix its number with '!', for example --port '!80'\n",
	)
	flags.String("instance-type", "nano", "Instance type")
	flags.Int64("min-scale", 1, "Min scale")
	flags.Int64("max-scale", 1, "Max scale")
	flags.StringSlice(
		"checks",
		[]string{""},
		"Update service healthchecks (available for services of type \"web\" only)\n"+
			"For HTTP healthchecks, use the format <PORT>:http:<PATH>, for example --checks 8080:http:/health\n"+
			"For TCP healthchecks, use the format <PORT>:tcp, for example --checks 8080:tcp\n"+
			"To delete a healthcheck, use !PORT, for example --checks '!8080'",
	)

	// Accept aliases: for example, allow user to use --port instead of --ports
	flags.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		aliases := map[string]string{
			"port":           "ports",
			"check":          "checks",
			"healthcheck":    "checks",
			"health-check":   "checks",
			"healthchecks":   "checks",
			"health-checks":  "checks",
			"route":          "routes",
			"region":         "regions",
			"git-docker-arg": "git-docker-args",
			"docker-arg":     "docker-args",
		}

		alias, exists := aliases[name]
		if exists {
			name = alias
		}
		return pflag.NormalizedName(name)
	})
}

func parseServiceDefinitionFlags(flags *pflag.FlagSet, definition *koyeb.DeploymentDefinition, useDefault bool) error {
	if useDefault || flags.Lookup("type").Changed {
		deploymentTypeStr, _ := flags.GetString("type")
		deploymentType, err := koyeb.NewDeploymentDefinitionTypeFromValue(strings.ToUpper(deploymentTypeStr))
		if err != nil {
			return &errors.CLIError{
				What: "Error while updating the service",
				Why:  "the --type flag is not valid",
				Additional: []string{
					"The --type flag must be either WEB or WORKER",
				},
				Orig:     nil,
				Solution: "Fix the --type flag and try again",
			}
		}
		definition.SetType(*deploymentType)
	}

	if definition.GetType() == koyeb.DEPLOYMENTDEFINITIONTYPE_WORKER {
		definition.Ports = nil
		definition.Routes = nil
		definition.HealthChecks = nil
	}

	if useDefault || flags.Lookup("env").Changed {
		envs, err := parseEnv(flags, definition.Env)
		if err != nil {
			return err
		}
		definition.SetEnv(envs)
	}

	if useDefault || flags.Lookup("instance-type").Changed {
		instanceType := koyeb.NewDeploymentInstanceTypeWithDefaults()
		val, _ := flags.GetString("instance-type")
		instanceType.SetType(val)
		definition.SetInstanceTypes([]koyeb.DeploymentInstanceType{*instanceType})
	}
	if useDefault || flags.Lookup("regions").Changed {
		regions, _ := flags.GetStringSlice("regions")
		definition.SetRegions(regions)
	}

	if useDefault && definition.GetType() == koyeb.DEPLOYMENTDEFINITIONTYPE_WEB || flags.Lookup("ports").Changed {
		ports, err := parsePorts(flags, definition.Ports)
		if err != nil {
			return err
		}
		definition.SetPorts(ports)
	}

	if useDefault && definition.GetType() == koyeb.DEPLOYMENTDEFINITIONTYPE_WEB || flags.Lookup("routes").Changed {
		routes, err := parseRoutes(flags, definition.Routes)
		if err != nil {
			return err
		}
		definition.SetRoutes(routes)
	}

	if useDefault || flags.Lookup("min-scale").Changed || flags.Lookup("max-scale").Changed {
		scaling := koyeb.NewDeploymentScalingWithDefaults()
		minScale, _ := flags.GetInt64("min-scale")
		maxScale, _ := flags.GetInt64("max-scale")
		scaling.SetMin(minScale)
		scaling.SetMax(maxScale)
		definition.SetScalings([]koyeb.DeploymentScaling{*scaling})
	}

	if flags.Lookup("checks").Changed {
		healthchecks, err := parseChecks(flags, definition.HealthChecks)
		if err != nil {
			return err
		}
		definition.SetHealthChecks(healthchecks)
	}

	// Docker
	if useDefault && !flags.Lookup("git").Changed || flags.Lookup("docker").Changed && !flags.Lookup("git").Changed {
		createDockerSource := koyeb.NewDockerSourceWithDefaults()
		image, _ := flags.GetString("docker")
		args, _ := flags.GetStringSlice("docker-args")
		command, _ := flags.GetString("docker-command")
		entrypoint, _ := flags.GetStringSlice("docker-entrypoint")
		image_registry_secret, _ := flags.GetString("docker-private-registry-secret")
		createDockerSource.SetImage(image)
		if command != "" {
			createDockerSource.SetCommand(command)
		}
		if image_registry_secret != "" {
			createDockerSource.SetImageRegistrySecret(image_registry_secret)
		}
		if len(args) > 0 {
			createDockerSource.SetArgs(args)
		}
		if len(entrypoint) > 0 {
			createDockerSource.SetEntrypoint(entrypoint)
		}
		definition.SetDocker(*createDockerSource)
		definition.Git = nil
	}
	// Git
	if flags.Lookup("git").Changed && !flags.Lookup("docker").Changed {
		builder, _ := flags.GetString("git-builder")
		if builder != "buildpack" && builder != "docker" {
			return &errors.CLIError{
				What: "Error while updating the service",
				Why:  "the --git-builder is invalid",
				Additional: []string{
					"The --git-builder must be either 'buildpack' or 'docker'",
				},
				Orig:     nil,
				Solution: "Fix the --git-builder and try again",
			}
		}

		if builder == "buildpack" && (flags.Lookup("git-docker-dockerfile").Changed ||
			flags.Lookup("git-docker-entrypoint").Changed ||
			flags.Lookup("git-docker-command").Changed ||
			flags.Lookup("git-docker-args").Changed ||
			flags.Lookup("git-docker-target").Changed) {
			return &errors.CLIError{
				What: "Error while updating the service",
				Why:  "invalid flag combination",
				Additional: []string{
					"The arguments --git-docker-* are used to configure the docker builder, and cannot be used with --git-builder=buildpack",
				},
				Orig:     nil,
				Solution: "Remove the --git-docker-* flags and try again, or use --git-builder=docker",
			}
		}

		if builder == "docker" && (flags.Lookup("git-buildpack-build-command").Changed ||
			flags.Lookup("git-buildpack-run-command").Changed) {
			return &errors.CLIError{
				What: "Error while updating the service",
				Why:  "invalid flag combination",
				Additional: []string{
					"The arguments --git-buildpack-* are used to configure the buildpack builder, and cannot be used with --git-builder=docker",
				},
				Orig:     nil,
				Solution: "Remove the --git-buildpack-* flags and try again, or use --git-builder=buildpack",
			}
		}

		createGitSource := koyeb.NewGitSourceWithDefaults()
		git, _ := flags.GetString("git")
		branch, _ := flags.GetString("git-branch")
		noDeployOnPush, _ := flags.GetBool("git-no-deploy-on-push")
		workdir, _ := flags.GetString("git-workdir")

		createGitSource.SetRepository(git)
		if branch != "" {
			createGitSource.SetBranch(branch)
		}
		createGitSource.SetNoDeployOnPush(noDeployOnPush)
		createGitSource.SetWorkdir(workdir)

		// Set builder
		switch builder {
		case "buildpack":
			// Legacy options for backward compatibility. We should use
			// --git-buildpack-build-command and --git-buildpack-run-command instead
			buildCommand, _ := flags.GetString("git-build-command")
			buildpackBuildCommand, _ := flags.GetString("git-buildpack-build-command")
			runCommand, _ := flags.GetString("git-run-command")
			buildpackRunCommand, _ := flags.GetString("git-buildpack-run-command")

			if buildCommand != "" && buildpackBuildCommand != "" {
				return &errors.CLIError{
					What: "Error while configuring the service",
					Why:  "can't use --git-build-command and --git-buildpack-build-command together",
					Additional: []string{
						"The command --git-build-command has been deprecated in favor of --git-buildpack-build-command.",
						"For backward compatibility, it is still possible to use --git-build-command, but it will be removed in a future release.",
						"In any case, the two options cannot be used together.",
					},
					Orig:     nil,
					Solution: "Only specify --git-buildpack-build-command",
				}
			}
			if runCommand != "" && buildpackRunCommand != "" {
				return &errors.CLIError{
					What: "Error while configuring the service",
					Why:  "can't use --git-run-command and --git-buildpack-run-command together",
					Additional: []string{
						"The command --git-run-command has been deprecated in favor of --git-buildpack-run-command.",
						"For backward compatibility, it is still possible to use --git-run-command, but it will be removed in a future release.",
						"In any case, the two options cannot be used together.",
					},
					Orig:     nil,
					Solution: "Only specify --git-buildpack-run-command",
				}
			}

			builder := koyeb.BuildpackBuilder{}
			if buildCommand != "" {
				builder.SetBuildCommand(buildCommand)
			} else if buildpackBuildCommand != "" {
				builder.SetBuildCommand(buildpackBuildCommand)
			}
			if runCommand != "" {
				builder.SetRunCommand(runCommand)
			} else if buildpackRunCommand != "" {
				builder.SetRunCommand(buildpackRunCommand)
			}
			createGitSource.SetBuildpack(builder)
		case "docker":
			dockerfile, _ := flags.GetString("git-docker-dockerfile")
			entrypoint, _ := flags.GetStringSlice("git-docker-entrypoint")
			command, _ := flags.GetString("git-docker-command")
			args, _ := flags.GetStringSlice("git-docker-args")
			target, _ := flags.GetString("git-docker-target")

			docker := koyeb.DockerBuilder{}
			if dockerfile != "" {
				docker.SetDockerfile(dockerfile)
			}
			if len(entrypoint) > 0 {
				docker.SetEntrypoint(entrypoint)
			}
			if command != "" {
				docker.SetCommand(command)
			}
			if len(args) > 0 {
				docker.SetArgs(args)
			}
			if target != "" {
				docker.SetTarget(target)
			}
			createGitSource.SetDocker(docker)
		}

		definition.SetGit(*createGitSource)
		definition.Docker = nil
	}
	return nil
}

// parseListFlags is the generic function parsing --env, --port, --routes and --checks.
// It gets the arguments given from the command line for the given flag, then
// builds a list of flags_list.Flag entries, and update the service
// configuration (given in existingItems) with the new values.
func parseListFlags[T any](
	flagName string,
	buildListFlags func([]string) ([]flags_list.Flag[T], error),
	flags *pflag.FlagSet,
	existingItems []T,
) ([]T, error) {
	values, err := flags.GetStringSlice(flagName)
	if err != nil {
		return nil, err
	}

	listFlags, err := buildListFlags(values)
	if err != nil {
		return nil, err
	}
	newItems := flags_list.ParseListFlags[T](listFlags, existingItems)
	return newItems, nil
}

// Parse --env flags
func parseEnv(flags *pflag.FlagSet, existingEnv []koyeb.DeploymentEnv) ([]koyeb.DeploymentEnv, error) {
	return parseListFlags("env", flags_list.NewEnvListFromFlags, flags, existingEnv)
}

// Parse --ports flags
func parsePorts(flags *pflag.FlagSet, existingPorts []koyeb.DeploymentPort) ([]koyeb.DeploymentPort, error) {
	return parseListFlags("ports", flags_list.NewPortListFromFlags, flags, existingPorts)
}

// Parse --routes flags
func parseRoutes(flags *pflag.FlagSet, existingRoutes []koyeb.DeploymentRoute) ([]koyeb.DeploymentRoute, error) {
	return parseListFlags("routes", flags_list.NewRouteListFromFlags, flags, existingRoutes)
}

// Parse --checks flags
func parseChecks(flags *pflag.FlagSet, existingHealthChecks []koyeb.DeploymentHealthCheck) ([]koyeb.DeploymentHealthCheck, error) {
	return parseListFlags("checks", flags_list.NewHealthcheckListFromFlags, flags, existingHealthChecks)
}
