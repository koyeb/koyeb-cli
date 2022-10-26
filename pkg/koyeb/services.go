package koyeb

import (
	"context"
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewServiceCmd() *cobra.Command {
	h := NewServiceHandler()

	serviceCmd := &cobra.Command{
		Use:               "services ACTION",
		Aliases:           []string{"s", "svc", "service"},
		Short:             "Services",
		PersistentPreRunE: h.InitHandler,
	}

	createServiceCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			createService := koyeb.NewCreateServiceWithDefaults()
			createDefinition := koyeb.NewDeploymentDefinitionWithDefaults()

			err := parseServiceDefinitionFlags(cmd.Flags(), createDefinition, true)
			if err != nil {
				return err
			}
			createDefinition.Name = koyeb.PtrString(args[0])

			createService.SetDefinition(*createDefinition)
			return h.Create(cmd, args, createService)
		},
	}
	addServiceDefinitionFlags(createServiceCmd.Flags())
	createServiceCmd.Flags().StringP("app", "a", "", "App")
	createServiceCmd.MarkFlagRequired("app")
	serviceCmd.AddCommand(createServiceCmd)

	getServiceCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get service",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Get,
	}
	serviceCmd.AddCommand(getServiceCmd)

	logsServiceCmd := &cobra.Command{
		Use:     "logs NAME",
		Aliases: []string{"l", "log"},
		Short:   "Get the service logs",
		Args:    cobra.ExactArgs(1),
		RunE:    h.Logs,
	}
	serviceCmd.AddCommand(logsServiceCmd)
	logsServiceCmd.Flags().String("instance", "", "Instance")
	logsServiceCmd.Flags().StringP("type", "t", "", "Type (runtime,build)")

	listServiceCmd := &cobra.Command{
		Use:   "list",
		Short: "List services",
		RunE:  h.List,
	}
	serviceCmd.AddCommand(listServiceCmd)
	listServiceCmd.Flags().StringP("app", "a", "", "App")
	listServiceCmd.Flags().StringP("name", "n", "", "Service name")

	describeServiceCmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe service",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Describe,
	}
	serviceCmd.AddCommand(describeServiceCmd)

	updateServiceCmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			updateService := koyeb.NewUpdateServiceWithDefaults()

			latestDeploy, resp, err := h.client.DeploymentsApi.ListDeployments(h.ctx).
				Limit("1").ServiceId(h.ResolveServiceArgs(args[0])).Execute()
			if err != nil {
				fatalApiError(err, resp)
			}
			if len(latestDeploy.GetDeployments()) == 0 {
				return errors.New("Unable to load latest deployment")
			}
			updateDef := latestDeploy.GetDeployments()[0].Definition
			err = parseServiceDefinitionFlags(cmd.Flags(), updateDef, false)
			if err != nil {
				return err
			}
			updateService.SetDefinition(*updateDef)
			return h.Update(cmd, args, updateService)
		},
	}
	addServiceDefinitionFlags(updateServiceCmd.Flags())
	serviceCmd.AddCommand(updateServiceCmd)

	redeployServiceCmd := &cobra.Command{
		Use:   "redeploy NAME",
		Short: "Redeploy service",
		Args:  cobra.ExactArgs(1),
		RunE:  h.ReDeploy,
	}
	serviceCmd.AddCommand(redeployServiceCmd)

	deleteServiceCmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete service",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Delete,
	}
	serviceCmd.AddCommand(deleteServiceCmd)

	pauseServiceCmd := &cobra.Command{
		Use:   "pause NAME",
		Short: "Pause service",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Pause,
	}
	serviceCmd.AddCommand(pauseServiceCmd)

	resumeServiceCmd := &cobra.Command{
		Use:   "resume NAME",
		Short: "Resume service",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Resume,
	}
	serviceCmd.AddCommand(resumeServiceCmd)

	return serviceCmd
}

func NewServiceHandler() *ServiceHandler {
	return &ServiceHandler{}
}

type ServiceHandler struct {
	ctx    context.Context
	client *koyeb.APIClient
	mapper *idmapper.Mapper
}

func (h *ServiceHandler) InitHandler(cmd *cobra.Command, args []string) error {
	h.ctx = getAuth(context.Background())
	h.client = getApiClient()
	h.mapper = idmapper.NewMapper(h.ctx, h.client)
	return nil
}

func (h *ServiceHandler) ResolveServiceArgs(val string) string {
	serviceMapper := h.mapper.Service()
	id, err := serviceMapper.ResolveID(val)
	if err != nil {
		fatalApiError(err, nil)
	}

	return id
}

func (h *ServiceHandler) ResolveAppArgs(val string) string {
	appMapper := h.mapper.App()
	id, err := appMapper.ResolveID(val)
	if err != nil {
		fatalApiError(err, nil)
	}

	return id
}

func addServiceDefinitionFlags(flags *pflag.FlagSet) {
	flags.String("git", "", "Git repository")
	flags.String("git-branch", "", "Git branch")
	flags.String("git-build-command", "", "Buid command")
	flags.String("git-run-command", "", "Run command")
	flags.Bool("git-no-deploy-on-push", false, "Disable new deployments creation when code changes are pushed on the configured branch")
	flags.String("docker", "", "Docker image")
	flags.String("docker-private-registry-secret", "", "Docker private registry secret")
	flags.String("docker-command", "", "Docker command")
	flags.StringSlice("docker-args", []string{}, "Docker args")
	flags.StringSlice("regions", []string{"fra"}, "Regions")
	flags.StringSlice("env", []string{}, "Env")
	flags.StringSlice("routes", []string{"/:80"}, "Ports")
	flags.StringSlice("ports", []string{"80:http"}, "Ports")
	flags.String("instance-type", "nano", "Instance type")
	flags.Int64("min-scale", 1, "Min scale")
	flags.Int64("max-scale", 1, "Max scale")
}

func parseServiceDefinitionFlags(flags *pflag.FlagSet, definition *koyeb.DeploymentDefinition, useDefault bool) error {

	if useDefault || flags.Lookup("env").Changed {
		env, _ := flags.GetStringSlice("env")
		envs := []koyeb.DeploymentEnv{}
		for _, e := range env {
			newEnv := koyeb.NewDeploymentEnvWithDefaults()

			split := strings.Split(e, "=")
			if len(split) < 2 {
				return errors.New("Unable to parse env")
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

	if useDefault || flags.Lookup("ports").Changed {
		port, _ := flags.GetStringSlice("ports")
		ports := []koyeb.DeploymentPort{}
		for _, p := range port {
			newPort := koyeb.NewDeploymentPortWithDefaults()

			split := strings.Split(p, ":")
			if len(split) < 1 {
				return errors.New("Unable to parse port")
			}
			portNum, err := strconv.Atoi(split[0])
			if err != nil {
				errors.Wrap(err, "Invalid port number")
			}
			newPort.Port = koyeb.PtrInt64(int64(portNum))
			newPort.Protocol = koyeb.PtrString("http")
			if len(split) > 1 {
				newPort.Protocol = koyeb.PtrString(split[1])
			}
			ports = append(ports, *newPort)

		}
		definition.SetPorts(ports)
	}

	if useDefault || flags.Lookup("routes").Changed {
		route, _ := flags.GetStringSlice("routes")
		routes := []koyeb.DeploymentRoute{}
		for _, p := range route {
			newRoute := koyeb.NewDeploymentRouteWithDefaults()

			spli := strings.Split(p, ":")
			if len(spli) < 1 {
				return errors.New("Unable to parse route")
			}
			newRoute.Path = koyeb.PtrString(spli[0])
			newRoute.Port = koyeb.PtrInt64(80)
			if len(spli) > 1 {
				portNum, err := strconv.Atoi(spli[1])
				if err != nil {
					errors.Wrap(err, "Invalid route number")
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

	// Docker
	if useDefault && !flags.Lookup("git").Changed || flags.Lookup("docker").Changed && !flags.Lookup("git").Changed {
		createDockerSource := koyeb.NewDockerSourceWithDefaults()
		image, _ := flags.GetString("docker")
		args, _ := flags.GetStringSlice("docker-args")
		command, _ := flags.GetString("docker-command")
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
		definition.SetDocker(*createDockerSource)
		definition.Git = nil
	}
	// Git
	if flags.Lookup("git").Changed && !flags.Lookup("docker").Changed {
		createGitSource := koyeb.NewGitSourceWithDefaults()
		git, _ := flags.GetString("git")
		branch, _ := flags.GetString("git-branch")
		buildCommand, _ := flags.GetString("git-build-command")
		runCommand, _ := flags.GetString("git-run-command")
		noDeployOnPush, _ := flags.GetBool("git-no-deploy-on-push")
		createGitSource.SetRepository(git)
		if branch != "" {
			createGitSource.SetBranch(branch)
		}

		if buildCommand != "" {
			createGitSource.SetBuildCommand(buildCommand)
		}

		if runCommand != "" {
			createGitSource.SetRunCommand(runCommand)
		}

		createGitSource.SetNoDeployOnPush(noDeployOnPush)
		definition.SetGit(*createGitSource)
		definition.Docker = nil
	}
	return nil
}
