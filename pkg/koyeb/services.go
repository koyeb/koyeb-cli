package koyeb

import (
	stderrors "errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	koyeb_errors "github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

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
	createServiceCmd.Flags().StringP("app", "a", "", "App")
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
				return &koyeb_errors.CLIError{
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
	flags.String("type", "WEB", `Service type, either "WEB" or "WORKER"`)
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
	flags.StringSlice("env", []string{}, "Environment variables, e.g. --env FOO=bar, or --env FOO=@bar to use the value of the secret bar")
	flags.StringSlice("routes", []string{"/:80"}, `Routes - Available for "WEB" service only`)
	flags.StringSlice("ports", []string{"80:http"}, `Ports - Available for "WEB" service only`)
	flags.String("instance-type", "nano", "Instance type")
	flags.Int64("min-scale", 1, "Min scale")
	flags.Int64("max-scale", 1, "Max scale")
	flags.StringSlice("checks", []string{""}, `HTTP healthcheck (<port>:http:<path>) and TCP healthcheck (<port>:tcp) - Available for "WEB" service only`)
}

func parseServiceDefinitionFlags(flags *pflag.FlagSet, definition *koyeb.DeploymentDefinition, useDefault bool) error {
	if useDefault || flags.Lookup("type").Changed {
		deploymentTypeStr, _ := flags.GetString("type")
		deploymentType, err := koyeb.NewDeploymentDefinitionTypeFromValue(deploymentTypeStr)
		if err != nil {
			return fmt.Errorf("%s is not a valid deployment type", deploymentTypeStr)
		}
		definition.SetType(*deploymentType)
	}

	if definition.GetType() == koyeb.DEPLOYMENTDEFINITIONTYPE_WORKER {
		definition.Ports = nil
		definition.Routes = nil
		definition.HealthChecks = nil
	}

	if useDefault || flags.Lookup("env").Changed {
		env, _ := flags.GetStringSlice("env")
		envs := []koyeb.DeploymentEnv{}
		for _, e := range env {
			newEnv := koyeb.NewDeploymentEnvWithDefaults()

			split := strings.SplitN(e, "=", 2)
			if len(split) != 2 || split[0] == "" || split[1] == "" {
				return &koyeb_errors.CLIError{
					What: "Error while configuring the service",
					Why:  fmt.Sprintf("unable to parse the environment variable \"%s\"", e),
					Additional: []string{
						"Environment variables must be specified as KEY=VALUE",
						"To use a secret as a value, specify KEY=@SECRET_NAME",
						"To specify an empty value, specify KEY=",
					},
					Orig:     nil,
					Solution: "Fix the environment variable and try again",
				}
			}

			newEnv.Key = koyeb.PtrString(split[0])
			if split[1][0] == '@' {
				newEnv.Secret = koyeb.PtrString(split[1][1:])
			} else {
				newEnv.Value = koyeb.PtrString(split[1])
			}

			envs = append(envs, *newEnv)
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
		port, _ := flags.GetStringSlice("ports")
		ports := []koyeb.DeploymentPort{}
		for _, p := range port {
			newPort := koyeb.NewDeploymentPortWithDefaults()

			split := strings.Split(p, ":")
			portNum, err := strconv.Atoi(split[0])
			if err != nil {
				return &koyeb_errors.CLIError{
					What: "Error while configuring the service",
					Why:  fmt.Sprintf("unable to parse the port \"%s\"", split[0]),
					Additional: []string{
						"Ports must be specified as PORT[:PROTOCOL]",
						"PORT must be a valid port number (e.g. 80)",
						"PROTOCOL must be either \"http\" or \"http2\". It can be omitted, in which case it defaults to \"http\"",
					},
					Orig:     nil,
					Solution: "Fix the port and try again",
				}
			}
			newPort.Port = koyeb.PtrInt64(int64(portNum))

			newPort.Protocol = koyeb.PtrString("http")
			if len(split) > 1 {
				if split[1] != "http" && split[1] != "http2" {
					return &koyeb_errors.CLIError{
						What: "Error while configuring the service",
						Why:  fmt.Sprintf("unable to parse the port protocol \"%s\"", split[1]),
						Additional: []string{
							"Ports must be specified as PORT[:PROTOCOL]",
							"PORT must be a valid port number (e.g. 80)",
							"PROTOCOL must be either \"http\" or \"http2\". It can be omitted, in which case it defaults to \"http\"",
						},
						Orig:     nil,
						Solution: "Fix the protocol and try again",
					}
				}
				newPort.Protocol = koyeb.PtrString(split[1])
			}
			ports = append(ports, *newPort)
		}
		definition.SetPorts(ports)
	}

	if useDefault && definition.GetType() == koyeb.DEPLOYMENTDEFINITIONTYPE_WEB || flags.Lookup("routes").Changed {
		route, _ := flags.GetStringSlice("routes")
		routes := []koyeb.DeploymentRoute{}
		for _, p := range route {
			newRoute := koyeb.NewDeploymentRouteWithDefaults()

			split := strings.Split(p, ":")
			if len(split) < 1 {
				return stderrors.New("Unable to parse route")
			}
			newRoute.Path = koyeb.PtrString(split[0])
			newRoute.Port = koyeb.PtrInt64(80)
			if len(split) > 1 {
				portNum, err := strconv.Atoi(split[1])
				if err != nil {
					return fmt.Errorf("invalid route number: %v", split[1])
				}
				newRoute.Port = koyeb.PtrInt64(int64(portNum))
			}
			routes = append(routes, *newRoute)

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
		checks, _ := flags.GetStringSlice("checks")
		healthchecks := []koyeb.DeploymentHealthCheck{}

		for _, c := range checks {
			healthcheck := koyeb.NewDeploymentHealthCheck()
			components := strings.Split(c, ":")
			componentsCount := len(components)
			if componentsCount < 2 || componentsCount > 3 {
				return fmt.Errorf(`Invalid checks: "%s", must be either "<port>:http:<path>" or "<port>:tcp"`, c)
			}

			healthcheckType := components[1]
			portStr := components[0]
			port, err := strconv.Atoi(portStr)
			if err != nil {
				return fmt.Errorf(`Invalid port: "%s"`, portStr)
			}

			switch healthcheckType {
			case "http":
				if componentsCount < 3 {
					return stderrors.New("Missing path definition for http check")
				}
				HTTPHealthCheck := koyeb.NewHTTPHealthCheck()
				HTTPHealthCheck.Port = koyeb.PtrInt64(int64(port))
				HTTPHealthCheck.Path = koyeb.PtrString(components[2])
				healthcheck.SetHttp(*HTTPHealthCheck)
			case "tcp":
				TCPHealthCheck := koyeb.NewTCPHealthCheck()
				TCPHealthCheck.Port = koyeb.PtrInt64(int64(port))
				healthcheck.SetTcp(*TCPHealthCheck)
			default:
				return fmt.Errorf(`Invalid healthcheck: "%s", must be either "http" or "tcp"`, healthcheckType)
			}
			healthchecks = append(healthchecks, *healthcheck)
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
			return stderrors.New("Invalid --git-builder, must be either 'buildpack' or 'docker'")
		}

		if builder == "buildpack" && (flags.Lookup("git-docker-dockerfile").Changed ||
			flags.Lookup("git-docker-entrypoint").Changed ||
			flags.Lookup("git-docker-command").Changed ||
			flags.Lookup("git-docker-args").Changed ||
			flags.Lookup("git-docker-target").Changed) {
			return stderrors.New(`Cannot use --git-docker-* options with --git-builder=buildpack`)
		}

		if builder == "docker" && (flags.Lookup("git-buildpack-build-command").Changed ||
			flags.Lookup("git-buildpack-run-command").Changed) {
			return stderrors.New(`Cannot use --git-buildpack-* options with --git-builder=docker`)
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
				return &koyeb_errors.CLIError{
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
				return &koyeb_errors.CLIError{
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
