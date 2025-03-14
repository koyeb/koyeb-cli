package koyeb

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/dates"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/flags_list"
	"github.com/sirupsen/logrus"
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
		Example: `
# Deploy a nginx docker image, listening on port 80
$> koyeb service create myservice --app myapp --docker nginx --port 80

# Deploy a nginx docker image and set the docker CMD explicitly, equivalent to docker CMD ["nginx", "-g", "daemon off;"]
$> koyeb service create myservice --app myapp --docker nginx --port 80 --docker-command nginx --docker-args '-g' --docker-args 'daemon off;'

# Build and deploy a GitHub repository using buildpack (default), set the environment variable PORT, and expose the port 9000 to the root route
$> koyeb service create myservice --app myapp --git github.com/koyeb/example-flask --git-branch main --env PORT=9000 --port 9000:http --route /:9000

# Build and deploy a GitHub repository using docker
$> koyeb service create myservice --app myapp --git github.com/org/name --git-branch main --git-builder docker

# Create a docker service, only accessible from the mesh (--route is not automatically created for TCP ports)
$> koyeb service create myservice --app myapp --docker nginx --port 80:tcp
`,
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			createService := koyeb.NewCreateServiceWithDefaults()
			createDefinition := koyeb.NewDeploymentDefinitionWithDefaults()

			err := h.parseServiceDefinitionFlags(ctx, cmd.Flags(), createDefinition)
			if err != nil {
				return err
			}

			serviceName, err := h.parseServiceNameWithoutApp(cmd, args[0])
			if err != nil {
				return err
			}

			createDefinition.Name = koyeb.PtrString(serviceName)
			createService.SetDefinition(*createDefinition)
			return h.Create(ctx, cmd, args, createService)
		}),
	}
	h.addServiceDefinitionFlags(createServiceCmd.Flags())
	createServiceCmd.Flags().StringP("app", "a", "", "Service application")
	serviceCmd.AddCommand(createServiceCmd)

	getServiceCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get service",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Get),
	}
	getServiceCmd.Flags().StringP("app", "a", "", "Service application")
	serviceCmd.AddCommand(getServiceCmd)

	unappliedChangesCmd := &cobra.Command{
		Use:   "unapplied-changes SERVICE_NAME",
		Short: "Show unapplied changes saved with the --save-only flag, which will be applied in the next deployment",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.ShowUnappliedChanges),
	}
	unappliedChangesCmd.Flags().StringP("app", "a", "", "Service application")
	serviceCmd.AddCommand(unappliedChangesCmd)

	var since dates.HumanFriendlyDate
	logsServiceCmd := &cobra.Command{
		Use:     "logs NAME",
		Aliases: []string{"l", "log"},
		Short:   "Get the service logs",
		Args:    cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			return h.Logs(ctx, cmd, since.Time, args)
		}),
	}
	logsServiceCmd.Flags().StringP("app", "a", "", "Service application")
	logsServiceCmd.Flags().String("instance", "", "Instance")
	logsServiceCmd.Flags().StringP("type", "t", "", "Type (runtime, build)")
	logsServiceCmd.Flags().Var(&since, "since", "DEPRECATED. Use --tail --start-time instead.")
	logsServiceCmd.Flags().Bool("tail", false, "Tail logs if no `--end-time` is provided.")
	logsServiceCmd.Flags().StringP("start-time", "s", "", "Return logs after this date")
	logsServiceCmd.Flags().StringP("end-time", "e", "", "Return logs before this date")
	logsServiceCmd.Flags().String("regex-search", "", "Filter logs returned with this regex")
	logsServiceCmd.Flags().String("text-search", "", "Filter logs returned with this text")
	logsServiceCmd.Flags().String("order", "asc", "Order logs by `asc` or `desc`")
	serviceCmd.AddCommand(logsServiceCmd)

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
	describeServiceCmd.Flags().StringP("app", "a", "", "Service application")
	serviceCmd.AddCommand(describeServiceCmd)

	execServiceCmd := &cobra.Command{
		Use:     "exec NAME CMD -- [args...]",
		Short:   "Run a command in the context of an instance selected among the service instances",
		Aliases: []string{"run", "attach"},
		Args:    cobra.MinimumNArgs(2),
		RunE:    WithCLIContext(h.Exec),
	}
	execServiceCmd.Flags().StringP("app", "a", "", "Service application")
	serviceCmd.AddCommand(execServiceCmd)

	updateServiceCmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update service",
		Args:  cobra.ExactArgs(1),
		Example: `
# Update the service "myservice" in the app "myapp", upsert the environment variable PORT and delete the environment variable DEBUG
$> koyeb service update myapp/myservice --env PORT=8001 --env '!DEBUG'

# Update the docker command of the service "myservice" in the app "myapp", equivalent to docker CMD ["nginx", "-g", "daemon off;"]
$> koyeb service update myapp/myservice --docker-command nginx --docker-args '-g' --docker-args 'daemon off;'

# Given a public service configured with the port 80:http and the route /:80, update it to make the service private, ie. only
# accessible from the mesh, by changing the port's protocol and removing the route
$> koyeb service update myapp/myservice --port 80:tcp --route '!/'
`,
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			serviceName, err := h.parseServiceName(cmd, args[0])
			if err != nil {
				return err
			}

			service, err := h.ResolveServiceArgs(ctx, serviceName)
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

			var updateDef *koyeb.DeploymentDefinition

			// If the --override flag is set, we start from a new deployment
			// definition with default values. Otherwise, we start from the
			// latest deployment definition.
			override, _ := cmd.Flags().GetBool("override")
			if override {
				updateDef = koyeb.NewDeploymentDefinitionWithDefaults()
				updateDef.Name = latestDeploy.GetDeployments()[0].Definition.Name
			} else {
				updateDef = latestDeploy.GetDeployments()[0].Definition
			}

			if updateDef.Git != nil && updateDef.Git.GetSha() != "" && !cmd.Flags().Lookup("git-sha").Changed {
				logrus.Warnf(
					"Warning: you are updating the service without specifying a commit with the --git-sha flag, and the service is currently deployed with the specific commit %s. If you want to deploy the latest commit of the branch instead, use --git-sha ''.",
					updateDef.Git.GetSha(),
				)
			}

			err = h.parseServiceDefinitionFlags(ctx, cmd.Flags(), updateDef)
			if err != nil {
				return err
			}
			updateService.SetDefinition(*updateDef)

			skipBuild, _ := cmd.Flags().GetBool("skip-build")
			updateService.SetSkipBuild(skipBuild)

			saveOnly, _ := cmd.Flags().GetBool("save-only")
			updateService.SetSaveOnly(saveOnly)

			return h.Update(ctx, cmd, args, updateService)
		}),
	}
	h.addServiceDefinitionFlags(updateServiceCmd.Flags())
	updateServiceCmd.Flags().StringP("app", "a", "", "Service application")
	updateServiceCmd.Flags().String("name", "", "Specify to update the service name")
	updateServiceCmd.Flags().Bool("override", false, "Override the service configuration with the new configuration instead of merging them")
	updateServiceCmd.Flags().Bool("skip-build", false, "If there has been at least one past successfully build deployment, use the last one instead of rebuilding. WARNING: this can lead to unexpected behavior if the build depends, for example, on environment variables.")
	updateServiceCmd.Flags().Bool("save-only", false, "Save the new configuration without deploying it")
	serviceCmd.AddCommand(updateServiceCmd)

	redeployServiceCmd := &cobra.Command{
		Use:   "redeploy NAME",
		Short: "Redeploy service",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.ReDeploy),
	}
	redeployServiceCmd.Flags().StringP("app", "a", "", "Service application")
	redeployServiceCmd.Flags().Bool("skip-build", false, "If there has been at least one past successfully build deployment, use the last one instead of rebuilding. WARNING: this can lead to unexpected behavior if the build depends, for example, on environment variables.")
	serviceCmd.AddCommand(redeployServiceCmd)
	redeployServiceCmd.Flags().Bool("use-cache", false, "Use cache to redeploy")

	deleteServiceCmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete service",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Delete),
	}
	deleteServiceCmd.Flags().StringP("app", "a", "", "Service application")
	serviceCmd.AddCommand(deleteServiceCmd)

	pauseServiceCmd := &cobra.Command{
		Use:   "pause NAME",
		Short: "Pause service",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Pause),
	}
	pauseServiceCmd.Flags().StringP("app", "a", "", "Service application")
	serviceCmd.AddCommand(pauseServiceCmd)

	resumeServiceCmd := &cobra.Command{
		Use:   "resume NAME",
		Short: "Resume service",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Resume),
	}
	resumeServiceCmd.Flags().StringP("app", "a", "", "Service application")
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

func (h *ServiceHandler) ResolveVolumeArgs(ctx *CLIContext, val string) (string, error) {
	volumeMapper := ctx.Mapper.Volume()
	id, err := volumeMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (h *ServiceHandler) addServiceDefinitionFlags(flags *pflag.FlagSet) {
	h.addServiceDefinitionFlagsForAllSources(flags)
	h.addServiceDefinitionFlagsForGitSource(flags)
	h.addServiceDefinitionFlagsForDockerSource(flags)
	h.addServiceDefinitionFlagsForArchiveSource(flags)
}

// Add the flags common to all sources: git, docker and archive
func (h *ServiceHandler) addServiceDefinitionFlagsForAllSources(flags *pflag.FlagSet) {
	// Global flags
	flags.String("type", "web", `Service type, either "web" or "worker"`)

	flags.StringSlice(
		"regions",
		[]string{},
		"Add a region where the service is deployed. You can specify this flag multiple times to deploy the service in multiple regions.\n"+
			"To update a service and remove a region, prefix the region name with '!', for example --region '!par'\n"+
			"If the region is not specified on service creation, the service is deployed in was\n",
	)
	flags.StringSlice(
		"env",
		[]string{},
		"Update service environment variables using the format KEY=VALUE, for example --env FOO=bar\n"+
			"To use the value of a secret as an environment variable, use the following syntax: --env FOO={{secret.bar}}\n"+
			"To delete an environment variable, prefix its name with '!', for example --env '!FOO'\n",
	)
	flags.String("instance-type", "nano", "Instance type")

	var strategy DeploymentStrategy
	flags.Var(&strategy, "deployment-strategy", `Deployment strategy, either "rolling" (default), "blue-green" or "immediate".`)

	flags.Int64("scale", 1, "Set both min-scale and max-scale")
	flags.Int64("min-scale", 1, "Min scale")
	flags.Int64("max-scale", 1, "Max scale")
	flags.Int64("autoscaling-average-cpu", 0, "Target CPU usage (in %) to trigger a scaling event. Set to 0 to disable CPU autoscaling.")
	flags.Int64("autoscaling-average-mem", 0, "Target memory usage (in %) to trigger a scaling event. Set to 0 to disable memory autoscaling.")
	flags.Int64("autoscaling-requests-per-second", 0, "Target requests per second to trigger a scaling event. Set to 0 to disable requests per second autoscaling.")
	flags.Int64("autoscaling-concurrent-requests", 0, "Target concurrent requests to trigger a scaling event. Set to 0 to disable concurrent requests autoscaling.")
	flags.Int64("autoscaling-requests-response-time", 0, "Target p95 response time to trigger a scaling event (in ms). Set to 0 to disable concurrent response time autoscaling.")
	flags.Bool("privileged", false, "Whether the service container should run in privileged mode")
	flags.Bool("skip-cache", false, "Whether to use the cache when building the service")

	// Global flags, only for services with the type "web" (not "worker")
	flags.StringSlice(
		"routes",
		nil,
		"Update service routes (available for services of type \"web\" only) using the format PATH[:PORT], for example '/foo:8080'\n"+
			"PORT defaults to 8000\n"+
			"To delete a route, use '!PATH', for example --route '!/foo'\n",
	)
	flags.StringSlice(
		"ports",
		nil,
		"Update service ports (available for services of type \"web\" only) using the format PORT[:PROTOCOL], for example --port 8080:http\n"+
			"PROTOCOL defaults to \"http\". Supported protocols are \"http\", \"http2\" and \"tcp\"\n"+
			"To delete an exposed port, prefix its number with '!', for example --port '!80'\n",
	)
	flags.StringSlice(
		"checks",
		nil,
		"Update service healthchecks (available for services of type \"web\" only)\n"+
			"For HTTP healthchecks, use the format <PORT>:http:<PATH>, for example --checks 8080:http:/health\n"+
			"For TCP healthchecks, use the format <PORT>:tcp, for example --checks 8080:tcp\n"+
			"To delete a healthcheck, use !PORT, for example --checks '!8080'\n",
	)
	flags.StringSlice(
		"checks-grace-period",
		nil,
		"Set healthcheck grace period in seconds.\n"+
			"Use the format <healthcheck>=<seconds>, for example --checks-grace-period 8080=10\n",
	)
	flags.StringSlice(
		"volumes",
		nil,
		"Update service volumes using the format VOLUME:PATH, for example --volume myvolume:/data."+
			"To delete a volume, use !VOLUME, for example --volume '!myvolume'\n",
	)
	flags.StringSlice(
		"config-file",
		nil,
		"Copy a local file to your service container using the format LOCAL_FILE:PATH:[PERMISSIONS]\n"+
			"for example --config-file /etc/data.yaml:/etc/data.yaml:0644\n"+
			"To delete a config file, use !PATH, for example --config-file !/etc/data.yaml\n",
	)

	// Configure aliases: for example, allow user to use --port instead of --ports
	flags.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		aliases := map[string]string{
			"port":  "ports",
			"check": "checks",

			"healthcheck":              "checks",
			"healthcheck-grace":        "checks-grace-period",
			"healthcheck-grace-period": "checks-grace-period",

			"health-check":              "checks",
			"health-check-grace":        "checks-grace-period",
			"health-check-grace-period": "checks-grace-period",

			"healthchecks":              "checks",
			"healthchecks-grace":        "checks-grace-period",
			"healthchecks-grace-period": "checks-grace-period",

			"health-checks":              "checks",
			"health-checks-graee":        "checks-grace-period",
			"health-checks-grace-period": "checks-grace-period",

			"strategy": "deployment-strategy",

			"route":              "routes",
			"volume":             "volumes",
			"region":             "regions",
			"git-docker-arg":     "git-docker-args",
			"docker-arg":         "docker-args",
			"archive-docker-arg": "archive-docker-args",
		}
		alias, exists := aliases[name]
		if exists {
			name = alias
		}
		return pflag.NormalizedName(name)
	})
}

// Add the flags for Git sources
func (h *ServiceHandler) addServiceDefinitionFlagsForGitSource(flags *pflag.FlagSet) {
	flags.String("git", "", "Git repository")
	flags.String("git-branch", "main", "Git branch")
	flags.String("git-sha", "", "Git commit SHA to deploy")
	flags.Bool("git-no-deploy-on-push", false, "Disable new deployments creation when code changes are pushed on the configured branch")
	flags.String("git-workdir", "", "Path to the sub-directory containing the code to build and deploy")
	flags.String("git-builder", "buildpack", `Builder to use, either "buildpack" (default) or "docker"`)

	// Git service: buildpack builder
	flags.String("git-build-command", "", "Buid command (legacy, prefer git-buildpack-build-command)")
	flags.String("git-run-command", "", "Run command (legacy, prefer git-buildpack-run-command)")
	flags.String("git-buildpack-build-command", "", "Buid command")
	flags.String("git-buildpack-run-command", "", "Run command")

	// Git service: docker builder
	flags.String("git-docker-dockerfile", "", "Dockerfile path")
	flags.StringSlice("git-docker-entrypoint", []string{}, "Docker entrypoint")
	flags.String("git-docker-command", "", "Set the docker CMD explicitly. To provide arguments to the command, use the --git-docker-args flag.")
	flags.StringSlice("git-docker-args", []string{}, "Set arguments to the docker command. To provide multiple arguments, use the --git-docker-args flag multiple times.")
	flags.String("git-docker-target", "", "Docker target")
}

// Add the flags for Docker sources
func (h *ServiceHandler) addServiceDefinitionFlagsForDockerSource(flags *pflag.FlagSet) {
	flags.String("docker", "", "Docker image")
	flags.String("docker-private-registry-secret", "", "Docker private registry secret")
	flags.StringSlice("docker-entrypoint", []string{}, "Docker entrypoint. To provide multiple arguments, use the --docker-entrypoint flag multiple times.")
	flags.String("docker-command", "", "Set the docker CMD explicitly. To provide arguments to the command, use the --docker-args flag.")
	flags.StringSlice("docker-args", []string{}, "Set arguments to the docker command. To provide multiple arguments, use the --docker-args flag multiple times.")
}

// Add the flags for Archive sources
func (h *ServiceHandler) addServiceDefinitionFlagsForArchiveSource(flags *pflag.FlagSet) {
	flags.String("archive", "", "Archive ID to deploy")
	flags.String("archive-builder", "buildpack", `Builder to use, either "buildpack" (default) or "docker"`)

	// Archive service: buildpack builder
	flags.String("archive-buildpack-build-command", "", "Buid command")
	flags.String("archive-buildpack-run-command", "", "Run command")

	// Archive service: docker builder
	flags.String("archive-docker-dockerfile", "", "Dockerfile path")
	flags.StringSlice("archive-docker-entrypoint", []string{}, "Docker entrypoint")
	flags.String("archive-docker-command", "", "Set the docker CMD explicitly. To provide arguments to the command, use the --archive-docker-args flag.")
	flags.StringSlice("archive-docker-args", []string{}, "Set arguments to the docker command. To provide multiple arguments, use the --archive-docker-args flag multiple times.")
	flags.String("archive-docker-target", "", "Docker target")
	flags.StringSlice("archive-ignore-dir", []string{".git", "node_modules", "vendor"},
		"Set directories to ignore when building the archive.\n"+
			"To ignore multiple directories, use the flag multiple times.\n"+
			"To include all directories, set the flag to an empty string.",
	)

}

func isFreeInstanceUsed(instanceTypes []koyeb.DeploymentInstanceType) bool {
	for _, instanceType := range instanceTypes {
		if instanceType.GetType() == "free" {
			return true
		}
	}
	return false
}

// parseServiceDefinitionFlags parses the flags related to the service definition, and updates the given definition accordingly.
func (h *ServiceHandler) parseServiceDefinitionFlags(ctx *CLIContext, flags *pflag.FlagSet, definition *koyeb.DeploymentDefinition) error {
	// For `koyeb service create`, the flag "name" does not exist so flags.Lookup("name") will return nil.
	// For `koyeb service update`, we only override the name in the definition if the flag is set.
	if flags.Lookup("name") != nil && flags.Lookup("name").Changed {
		name, _ := flags.GetString("name")
		definition.SetName(name)
	}

	type_, err := h.parseType(flags, definition.GetType())
	if err != nil {
		return err
	}
	definition.SetType(type_)

	strategy, err := h.parseDeploymentStrategy(flags, definition.GetStrategy())
	if err != nil {
		return err
	}
	definition.SetStrategy(strategy)

	skipCache, _ := flags.GetBool("skip-cache")
	definition.SetSkipCache(skipCache)

	envs, err := h.parseEnv(flags, definition.Env)
	if err != nil {
		return err
	}
	definition.SetEnv(envs)

	definition.SetInstanceTypes(h.parseInstanceType(flags, definition.GetInstanceTypes()))

	ports, err := h.parsePorts(definition.GetType(), flags, definition.Ports)
	if err != nil {
		return err
	}
	definition.SetPorts(ports)

	routes, err := h.parseRoutes(definition.GetType(), flags, definition.Routes)
	if err != nil {
		return err
	}
	definition.SetRoutes(routes)

	if definition.GetType() == koyeb.DEPLOYMENTDEFINITIONTYPE_WEB {
		err = h.setDefaultPortsAndRoutes(definition, definition.Ports, definition.Routes)
		if err != nil {
			return err
		}
	}

	isFreeUsed := isFreeInstanceUsed(definition.GetInstanceTypes())
	definition.SetScalings(h.parseScalings(isFreeUsed, flags, definition.Scalings))

	healthchecks, err := h.parseChecks(definition.GetType(), flags, definition.HealthChecks)
	if err != nil {
		return err
	}
	definition.SetHealthChecks(healthchecks)

	regions, err := h.parseRegions(flags, definition.GetRegions())
	if err != nil {
		return err
	}
	if flags.Lookup("regions").Changed && len(regions) >= 2 {
		logrus.Warnf(
			"Attention: you are deploying your service in %d regions (%s) which may impact your billing. If you intended to deploy your service in only one region, remove the regions you don't want to deploy to with `koyeb service update <app>/<service> --region '!<region>'`.",
			len(regions),
			strings.Join(regions, ", "),
		)
	}
	// Scalings and environment variables refer to regions, so we must call setRegions after definition.SetScalings and definition.SetEnv.
	h.setRegions(definition, regions)

	err = h.setSource(ctx, definition, flags)
	if err != nil {
		return err
	}

	volumes, err := h.parseVolumes(ctx, flags, definition.Volumes)
	if err != nil {
		return err
	}
	definition.SetVolumes(volumes)

	files, err := h.parseConfigFiles(ctx, flags, definition.ConfigFiles)
	if err != nil {
		return err
	}
	definition.SetConfigFiles(files)

	return nil
}

// Parse --type
func (h *ServiceHandler) parseType(flags *pflag.FlagSet, currentType koyeb.DeploymentDefinitionType) (koyeb.DeploymentDefinitionType, error) {
	if !flags.Lookup("type").Changed {
		// New service: return the default value
		if currentType == koyeb.DEPLOYMENTDEFINITIONTYPE_INVALID {
			def := strings.ToUpper(flags.Lookup("type").DefValue)
			ret, err := koyeb.NewDeploymentDefinitionTypeFromValue(def)
			if err != nil {
				panic(err)
			}
			return *ret, nil
		}
		// Existing service: return the type currently configured
		return currentType, nil
	}

	value, _ := flags.GetString("type")
	ret, err := koyeb.NewDeploymentDefinitionTypeFromValue(strings.ToUpper(value))
	if err != nil {
		return "", &errors.CLIError{
			What: "Error while updating the service",
			Why:  "the --type flag is not valid",
			Additional: []string{
				"The --type flag must be either \"web\" or \"worker\"",
			},
			Orig:     nil,
			Solution: "Fix the --type flag and try again",
		}
	}
	return *ret, nil
}

// Parse --instance-type
func (h *ServiceHandler) parseInstanceType(flags *pflag.FlagSet, currentInstanceTypes []koyeb.DeploymentInstanceType) []koyeb.DeploymentInstanceType {
	if !flags.Lookup("instance-type").Changed {
		// New service: return the default value
		if len(currentInstanceTypes) == 0 {
			ret := koyeb.NewDeploymentInstanceTypeWithDefaults()
			ret.SetType(flags.Lookup("instance-type").DefValue)
			return []koyeb.DeploymentInstanceType{*ret}
		}
		// Existing service
		return currentInstanceTypes
	}
	ret := koyeb.NewDeploymentInstanceTypeWithDefaults()
	value, _ := flags.GetString("instance-type")
	ret.SetType(value)
	return []koyeb.DeploymentInstanceType{*ret}
}

// Parse --deployment-strategy
func (h *ServiceHandler) parseDeploymentStrategy(flags *pflag.FlagSet, currentStrategy koyeb.DeploymentStrategy) (koyeb.DeploymentStrategy, error) {
	if !flags.Lookup("deployment-strategy").Changed {
		return currentStrategy, nil
	}
	flagValue := flags.Lookup("deployment-strategy").Value.(*DeploymentStrategy)
	strategy := koyeb.DeploymentStrategyType(*flagValue)
	return koyeb.DeploymentStrategy{
		Type: &strategy,
	}, nil
}

// parseListFlags is the generic function parsing --env, --port, --routes, --checks, --regions and --volumes
// It gets the arguments given from the command line for the given flag, then
// builds a list of flags_list.Flag entries, and update the service
// configuration (given in existingItems) with the new values.
func parseListFlags[T any](
	flagName string,
	buildListFlags func([]string) ([]flags_list.Flag[T], error),
	flags *pflag.FlagSet,
	currentItems []T,
) ([]T, error) {
	values, err := flags.GetStringSlice(flagName)
	if err != nil {
		return nil, err
	}

	listFlags, err := buildListFlags(values)
	if err != nil {
		return nil, err
	}
	newItems := flags_list.ParseListFlags(listFlags, currentItems)
	return newItems, nil
}

// Parse --env
func (h *ServiceHandler) parseEnv(flags *pflag.FlagSet, currentEnv []koyeb.DeploymentEnv) ([]koyeb.DeploymentEnv, error) {
	return parseListFlags("env", flags_list.NewEnvListFromFlags, flags, currentEnv)
}

// Parse --ports
func (h *ServiceHandler) parsePorts(type_ koyeb.DeploymentDefinitionType, flags *pflag.FlagSet, currentPorts []koyeb.DeploymentPort) ([]koyeb.DeploymentPort, error) {
	newPorts, err := parseListFlags("ports", flags_list.NewPortListFromFlags, flags, currentPorts)
	if err != nil {
		return nil, err
	}
	if len(newPorts) > 0 && type_ != koyeb.DEPLOYMENTDEFINITIONTYPE_WEB {
		errmsg := ""
		for _, port := range newPorts {
			errmsg = fmt.Sprintf("%s --port '!%d'", errmsg, port.GetPort())
		}
		return nil, &errors.CLIError{
			What: "Error while configuring the service",
			Why:  `your service has ports configured, which is only possible for services of type "web"`,
			Additional: []string{
				`To change the type of your service, set --type to "web".`,
				"To remove all the ports from your service, add the following flags:",
				errmsg,
			},
			Orig:     nil,
			Solution: "Fix the service type or remove the ports from your service, and try again",
		}
	}
	return newPorts, nil
}

// Parse --routes
func (h *ServiceHandler) parseRoutes(type_ koyeb.DeploymentDefinitionType, flags *pflag.FlagSet, currentRoutes []koyeb.DeploymentRoute) ([]koyeb.DeploymentRoute, error) {
	newRoutes, err := parseListFlags("routes", flags_list.NewRouteListFromFlags, flags, currentRoutes)
	if err != nil {
		return nil, err
	}
	if len(newRoutes) > 0 && type_ != koyeb.DEPLOYMENTDEFINITIONTYPE_WEB {
		errmsg := ""
		for _, route := range newRoutes {
			errmsg = fmt.Sprintf("%s --route '!%s'", errmsg, route.GetPath())
		}
		return nil, &errors.CLIError{
			What: "Error while configuring the service",
			Why:  `your service has routes configured, which is only possible for services of type "web"`,
			Additional: []string{
				`To change the type of your service, set --type to "web".`,
				"To remove all the routes from your service, add the following flags:",
				errmsg,
			},
			Orig:     nil,
			Solution: "Fix the service type or remove the routes from your service, and try again",
		}
	}
	return newRoutes, nil
}

// Set default port to `portNumber` and `http`
func (h *ServiceHandler) getDeploymentPort(portNumber int64) []koyeb.DeploymentPort {
	newPort := koyeb.NewDeploymentPortWithDefaults()
	newPort.SetPort(portNumber)
	newPort.SetProtocol("http")
	return []koyeb.DeploymentPort{*newPort}
}

// Set default route to `portNumber` and `/`
func (h *ServiceHandler) getDeploymentRoute(portNumber int64) []koyeb.DeploymentRoute {
	newRoute := koyeb.NewDeploymentRouteWithDefaults()
	newRoute.SetPath("/")
	newRoute.SetPort(portNumber)
	return []koyeb.DeploymentRoute{*newRoute}
}

// Dynamically sets the defaults ports and routes for "web" services
func (h *ServiceHandler) setDefaultPortsAndRoutes(definition *koyeb.DeploymentDefinition, currentPorts []koyeb.DeploymentPort, currentRoutes []koyeb.DeploymentRoute) error {
	switch {
	// If no route and no port is specified, add the default route and port
	case len(currentPorts) == 0 && len(currentRoutes) == 0:
		definition.SetPorts(h.getDeploymentPort(8000))
		definition.SetRoutes(h.getDeploymentRoute(8000))

	// When one or more ports are set but no route is explicitly configured:
	// - if only one HTTP/HTTP2 port is defined, create the default route using that port
	// - if more than one HTTP/HTTP2 port is set, we return an error as we can't determine routes configuration
	case len(currentPorts) > 0 && len(currentRoutes) == 0:
		httpPorts := []koyeb.DeploymentPort{}

		for _, port := range currentPorts {
			if port.GetProtocol() == "http" || port.GetProtocol() == "http2" {
				httpPorts = append(httpPorts, port)
			}
		}

		if len(httpPorts) == 1 {
			definition.SetRoutes(h.getDeploymentRoute(httpPorts[0].GetPort()))
		}

		if len(httpPorts) > 1 {
			return &errors.CLIError{
				What: "Error while configuring the service",
				Why:  `your service has two or more HTTP/HTTP2 ports set but no matching routes`,
				Additional: []string{
					"For each  HTTP/HTTP2 port, you must specify a matching route with the --routes flag",
				},
				Orig:     nil,
				Solution: "Set the routes and try again",
			}
		}

	// If one or more routes are set but no port is set:
	// - if more than one route is set, we can't determine which one should be used to create the default port, so we return an error
	// - if exactly only one route is set, we create the default port with this the port value of the route
	case len(currentRoutes) > 0 && len(currentPorts) == 0:
		if len(currentRoutes) > 1 {
			return &errors.CLIError{
				What: "Error while configuring the service",
				Why:  `your service has two or more routes set but no matching ports`,
				Additional: []string{
					"For each route, you must specify a matching port with the --ports flag",
				},
				Orig:     nil,
				Solution: "Set the ports and try again",
			}
		}
		portNumber := currentRoutes[0].GetPort()
		definition.SetPorts(h.getDeploymentPort(portNumber))
	}
	return nil
}

// Parse --checks
func (h *ServiceHandler) parseChecks(type_ koyeb.DeploymentDefinitionType, flags *pflag.FlagSet, currentHealthChecks []koyeb.DeploymentHealthCheck) ([]koyeb.DeploymentHealthCheck, error) {
	newChecks, err := parseListFlags("checks", flags_list.NewHealthcheckListFromFlags, flags, currentHealthChecks)
	if err != nil {
		return nil, err
	}
	if len(newChecks) > 0 && type_ != koyeb.DEPLOYMENTDEFINITIONTYPE_WEB {
		errmsg := ""
		for _, check := range newChecks {
			if check.HasHttp() {
				errmsg = fmt.Sprintf("%s --check '!%d'", errmsg, *check.GetHttp().Port)
			} else {
				errmsg = fmt.Sprintf("%s --check '!%d'", errmsg, *check.GetTcp().Port)
			}
		}
		return nil, &errors.CLIError{
			What: "Error while configuring the service",
			Why:  `--checks can only be specified for "web" services`,
			Additional: []string{
				`To change the type of your service, set --type to "web".`,
				"To remove all the healthchecks from your service, add the following flags:",
				errmsg,
			},
			Orig:     nil,
			Solution: "Fix the service type or remove the healthchecks from your service, and try again",
		}
	}

	checksGracePeriod, _ := flags.GetStringSlice("checks-grace-period")
	for _, val := range checksGracePeriod {
		if err := h.parseChecksGracePeriod(newChecks, val); err != nil {
			return nil, err
		}
	}
	return newChecks, nil
}

// parseChecksGracePeriod parses the --checks-grace-period flag and updates the healthchecks with the specified grace period.
func (h *ServiceHandler) parseChecksGracePeriod(checks []koyeb.DeploymentHealthCheck, grace string) error {
	parts := strings.Split(grace, "=")
	if len(parts) != 2 {
		return &errors.CLIError{
			What:       "Invalid grace period",
			Why:        "--checks-grace-period should be formatted as <healthcheck port number>=<grace period in seconds>",
			Additional: nil,
			Orig:       nil,
			Solution:   "Provide a valid grace period and try again",
		}
	}

	port, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return &errors.CLIError{
			What:       "Invalid grace period",
			Why:        "the grace period should be formatted as <healthcheck port number>=<grace period in seconds>",
			Additional: nil,
			Orig:       nil,
			Solution:   "Provide a valid grace period and try again",
		}
	}

	graceValue, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return &errors.CLIError{
			What:       "Invalid grace period",
			Why:        fmt.Sprintf("the grace period should be a number of seconds, not %s", parts[1]),
			Additional: nil,
			Orig:       nil,
			Solution:   "Provide a valid grace period and try again",
		}
	}

	for idx := range checks {
		if (checks[idx].HasHttp() && *checks[idx].GetHttp().Port == port) ||
			(checks[idx].HasTcp() && *checks[idx].GetTcp().Port == port) {
			checks[idx].GracePeriod = &graceValue
			return nil
		}
	}
	return &errors.CLIError{
		What: "Invalid grace period",
		Why:  "--checks-grace-period does not match any healthcheck",
		Additional: []string{
			fmt.Sprintf("The flag --checks-grace-period has been specified for the healthcheck on port %d, but no healthcheck with this port has been configured", port),
		},
		Orig:     nil,
		Solution: "Fix the flag --checks-grace-period to match an existing healthcheck or remove the grace period, and try again",
	}
}

// Parse --regions
func (h *ServiceHandler) parseRegions(flags *pflag.FlagSet, currentRegions []string) ([]string, error) {
	newRegions, err := parseListFlags("regions", flags_list.NewRegionsListFromFlags, flags, currentRegions)
	if err != nil {
		return nil, err
	}
	// For new services, if no region is specified, add the default region
	if !flags.Lookup("regions").Changed && len(currentRegions) == 0 {
		newRegions = []string{"was"}
	}
	return newRegions, nil
}

// Parse --min-scale and --max-scale
func (h *ServiceHandler) parseScalings(isFreeUsed bool, flags *pflag.FlagSet, currentScalings []koyeb.DeploymentScaling) []koyeb.DeploymentScaling {
	var minScale, maxScale int64

	if flags.Lookup("min-scale").Changed {
		minScale, _ = flags.GetInt64("min-scale")
	} else if flags.Lookup("scale").Changed || !isFreeUsed {
		minScale, _ = flags.GetInt64("scale")
	} else {
		minScale = 0
	}

	if flags.Lookup("max-scale").Changed {
		maxScale, _ = flags.GetInt64("max-scale")
	} else {
		maxScale, _ = flags.GetInt64("scale")
	}

	// If there is no scaling configured, return the default values
	if len(currentScalings) == 0 {
		scaling := koyeb.NewDeploymentScalingWithDefaults()
		scaling.SetMin(minScale)
		scaling.SetMax(maxScale)
		h.setScalingsTargets(flags, scaling)
		return []koyeb.DeploymentScaling{*scaling}
	} else {
		// Otherwise, update the current scaling configuration only if one of the scale flags has been provided
		for idx := range currentScalings {
			if flags.Lookup("scale").Changed || flags.Lookup("min-scale").Changed {
				currentScalings[idx].SetMin(minScale)
			}
			if flags.Lookup("scale").Changed || flags.Lookup("max-scale").Changed {
				currentScalings[idx].SetMax(maxScale)
			}
			h.setScalingsTargets(flags, &currentScalings[idx])
		}
	}
	return currentScalings
}

// setScalingsTargets updates the scaling targets in a koyeb.DeploymentScaling object based on the specified flags.
// It checks for changes in flags related to autoscaling parameters (CPU, memory, requests per second, concurrent requests, response time).
// If a change is detected in any of these flags, the function either updates an existing scaling target or creates a new one.
//
// setScalingsTargets performs the following steps:
//  1. For each autoscaling parameter (CPU, memory, requests per second,
//     concurrent requests, response time), it checks if the corresponding flag has
//     changed.
//  2. If a flag has changed, it iterates over the existing targets to find a
//     matching target type.
//  3. If a matching target is found, it updates the target with the new value or
//     removes the target if the value is 0.
//  4. If no matching target is found, it creates a new target with the specified
//     value and appends it to the scaling targets.
//
// Note: there is no way to easily avoid the code duplication in this function
// because NewDeploymentScalingTarget{AverageCPU,AverageMem,RequestsPerSecond,ConcurrentRequests,...}
// do not implement a common interface.
func (h *ServiceHandler) setScalingsTargets(flags *pflag.FlagSet, scaling *koyeb.DeploymentScaling) {
	if scaling.Targets == nil || scaling.GetMin() == scaling.GetMax() {
		scaling.Targets = []koyeb.DeploymentScalingTarget{}
	}

	if flags.Lookup("autoscaling-average-cpu").Changed {
		value, _ := flags.GetInt64("autoscaling-average-cpu")
		newTargets := []koyeb.DeploymentScalingTarget{}
		found := false

		for _, target := range scaling.GetTargets() {
			// Update the current target if it is a CPU target, or remove it if value is 0.
			if target.HasAverageCpu() {
				found = true
				if value == 0 {
					continue
				}
				cpu := koyeb.NewDeploymentScalingTargetAverageCPU()
				cpu.SetValue(value)
				target.SetAverageCpu(*cpu)
				newTargets = append(newTargets, target)
			} else {
				newTargets = append(newTargets, target)
			}
		}
		// No existing CPU target found, create a new entry
		if !found && value > 0 {
			target := koyeb.NewDeploymentScalingTarget()
			cpu := koyeb.NewDeploymentScalingTargetAverageCPU()
			cpu.SetValue(value)
			target.SetAverageCpu(*cpu)
			newTargets = append(newTargets, *target)
		}
		scaling.Targets = newTargets
	}

	if flags.Lookup("autoscaling-average-mem").Changed {
		value, _ := flags.GetInt64("autoscaling-average-mem")
		newTargets := []koyeb.DeploymentScalingTarget{}
		found := false

		for _, target := range scaling.GetTargets() {
			if target.HasAverageMem() {
				found = true
				if value == 0 {
					continue
				}
				mem := koyeb.NewDeploymentScalingTargetAverageMem()
				mem.SetValue(value)
				target.SetAverageMem(*mem)
				newTargets = append(newTargets, target)
			} else {
				newTargets = append(newTargets, target)
			}
		}
		if !found && value > 0 {
			target := koyeb.NewDeploymentScalingTarget()
			mem := koyeb.NewDeploymentScalingTargetAverageMem()
			mem.SetValue(value)
			target.SetAverageMem(*mem)
			newTargets = append(newTargets, *target)
		}
		scaling.Targets = newTargets
	}

	if flags.Lookup("autoscaling-requests-per-second").Changed {
		value, _ := flags.GetInt64("autoscaling-requests-per-second")
		newTargets := []koyeb.DeploymentScalingTarget{}
		found := false

		for _, target := range scaling.GetTargets() {
			if target.HasRequestsPerSecond() {
				found = true
				if value == 0 {
					continue
				}
				rps := koyeb.NewDeploymentScalingTargetRequestsPerSecond()
				rps.SetValue(value)
				target.SetRequestsPerSecond(*rps)
				newTargets = append(newTargets, target)
			} else {
				newTargets = append(newTargets, target)
			}
		}
		if !found && value > 0 {
			target := koyeb.NewDeploymentScalingTarget()
			rps := koyeb.NewDeploymentScalingTargetRequestsPerSecond()
			rps.SetValue(value)
			target.SetRequestsPerSecond(*rps)
			newTargets = append(newTargets, *target)
		}
		scaling.Targets = newTargets
	}

	if flags.Lookup("autoscaling-concurrent-requests").Changed {
		value, _ := flags.GetInt64("autoscaling-concurrent-requests")
		newTargets := []koyeb.DeploymentScalingTarget{}
		found := false

		for _, target := range scaling.GetTargets() {
			if target.HasConcurrentRequests() {
				found = true
				if value == 0 {
					continue
				}
				cr := koyeb.NewDeploymentScalingTargetConcurrentRequests()
				cr.SetValue(value)
				target.SetConcurrentRequests(*cr)
				newTargets = append(newTargets, target)
			} else {
				newTargets = append(newTargets, target)
			}
		}
		if !found && value > 0 {
			target := koyeb.NewDeploymentScalingTarget()
			cr := koyeb.NewDeploymentScalingTargetConcurrentRequests()
			cr.SetValue(value)
			target.SetConcurrentRequests(*cr)
			newTargets = append(newTargets, *target)
		}
		scaling.Targets = newTargets
	}

	if flags.Lookup("autoscaling-requests-response-time").Changed {
		value, _ := flags.GetInt64("autoscaling-requests-response-time")
		newTargets := []koyeb.DeploymentScalingTarget{}
		found := false

		for _, target := range scaling.GetTargets() {
			if target.HasRequestsResponseTime() {
				found = true
				if value == 0 {
					continue
				}
				cr := koyeb.NewDeploymentScalingTargetRequestsResponseTime()
				cr.SetValue(value)
				// For now, we hardcode the quantile to 95. In the future, we may want to expose it as a flag.
				cr.SetQuantile(95)
				target.SetRequestsResponseTime(*cr)
				newTargets = append(newTargets, target)
			} else {
				newTargets = append(newTargets, target)
			}
		}
		if !found && value > 0 {
			target := koyeb.NewDeploymentScalingTarget()
			cr := koyeb.NewDeploymentScalingTargetRequestsResponseTime()
			cr.SetValue(value)
			cr.SetQuantile(95)
			target.SetRequestsResponseTime(*cr)
			newTargets = append(newTargets, *target)
		}
		scaling.Targets = newTargets
	}
}

// Parse --git-* and --docker-* flags to set deployment.Git or deployment.Docker
func (h *ServiceHandler) setSource(ctx *CLIContext, definition *koyeb.DeploymentDefinition, flags *pflag.FlagSet) error {
	hasGitFlags := false
	hasDockerFlags := false
	hasArchiveFlags := false
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Changed {
			if strings.HasPrefix(flag.Name, "git") {
				hasGitFlags = true
			} else if strings.HasPrefix(flag.Name, "docker") {
				hasDockerFlags = true
			} else if strings.HasPrefix(flag.Name, "archive") {
				hasArchiveFlags = true
			}
		}
	})
	if hasGitFlags && hasDockerFlags || hasGitFlags && hasArchiveFlags || hasDockerFlags && hasArchiveFlags {
		return &errors.CLIError{
			What: "Error while updating the service",
			Why:  "invalid flag combination",
			Additional: []string{
				"Your service has conflicting options --git*, --docker* and/or --archive*.",
				"To build a GitHub repository, specify --git github.com/<owner>/<repo>",
				"To deploy a Docker image hosted on the Docker Hub or a private registry, specify --docker <image>",
				"To deploy from an archive created with `koyeb archive create`, specify --archive <archive-id>",
			},
			Orig:     nil,
			Solution: "Fix the flags and try again",
		}
	}
	if hasDockerFlags {
		// If --docker-* flags are set and the service already has a Docker
		// source, update it.
		docker := definition.GetDocker()
		source, err := h.parseDockerSource(ctx, flags, &docker)
		if err != nil {
			return err
		}
		definition.SetDocker(*source)
		definition.Git = nil
		definition.Archive = nil
		// If --docker-* flags are set and the service has a Git source, replace
		// it with a Docker source.
	} else if hasGitFlags {
		git := definition.GetGit()
		source, err := h.parseGitSource(flags, &git)
		if err != nil {
			return err
		}
		definition.SetGit(*source)
		definition.Docker = nil
		definition.Archive = nil
	} else if hasArchiveFlags {
		archive := definition.GetArchive()
		source, err := h.parseArchiveSource(flags, &archive)
		if err != nil {
			return err
		}
		definition.SetArchive(*source)
		definition.Git = nil
		definition.Docker = nil
	} else if definition.HasDocker() {
		// If none of the flags --git, --docker and --archive are set, parse the
		// flag to update the existing Docker source. This might seem to be a
		// no-op (since the docker flags are not set, you could expect the
		// source to remain the same), but it is necessary to update the
		// --privileged flag, for example.
		docker := definition.GetDocker()
		source, err := h.parseDockerSource(ctx, flags, &docker)
		if err != nil {
			return err
		}
		definition.SetDocker(*source)
		definition.Git = nil
		definition.Archive = nil
	} else if definition.HasArchive() {
		archive := definition.GetArchive()
		source, err := h.parseArchiveSource(flags, &archive)
		if err != nil {
			return err
		}
		definition.SetArchive(*source)
		definition.Git = nil
		definition.Docker = nil
	} else {
		// Same as above, but for the Git source.
		git := definition.GetGit()
		source, err := h.parseGitSource(flags, &git)
		if err != nil {
			return err
		}
		definition.SetGit(*source)
		definition.Docker = nil
		definition.Archive = nil
	}
	return nil
}

// Parse --docker-* flags
func (h *ServiceHandler) parseDockerSource(ctx *CLIContext, flags *pflag.FlagSet, source *koyeb.DockerSource) (*koyeb.DockerSource, error) {
	// docker-private-registry-secret needs to be parsed first, because checkDockerImage reads it
	if flags.Lookup("docker-private-registry-secret").Changed {
		secret, _ := flags.GetString("docker-private-registry-secret")
		source.SetImageRegistrySecret(secret)
	}
	if flags.Lookup("docker").Changed {
		image, _ := flags.GetString("docker")
		source.SetImage(image)
		if err := h.checkDockerImage(ctx, source); err != nil {
			return nil, err
		}
	}
	if flags.Lookup("docker-args").Changed {
		args, _ := flags.GetStringSlice("docker-args")
		source.SetArgs(args)
	}
	if flags.Lookup("docker-command").Changed {
		command, _ := flags.GetString("docker-command")
		source.SetCommand(command)
	}
	if flags.Lookup("docker-entrypoint").Changed {
		entrypoint, _ := flags.GetStringSlice("docker-entrypoint")
		source.SetEntrypoint(entrypoint)
	}
	if flags.Lookup("privileged").Changed {
		privileged, _ := flags.GetBool("privileged")
		source.SetPrivileged(privileged)
	}
	return source, nil
}

// Parse --git-* flags
func (h *ServiceHandler) parseGitSource(flags *pflag.FlagSet, source *koyeb.GitSource) (*koyeb.GitSource, error) {
	if flags.Lookup("git").Changed {
		repository, _ := flags.GetString("git")
		source.SetRepository(repository)
	}
	if source.GetBranch() == "" || flags.Lookup("git-branch").Changed {
		branch, _ := flags.GetString("git-branch")
		source.SetBranch(branch)
	}
	if flags.Lookup("git-sha").Changed {
		sha, _ := flags.GetString("git-sha")
		source.SetSha(sha)
	}
	if flags.Lookup("git-no-deploy-on-push").Changed {
		noDeployOnPush, _ := flags.GetBool("git-no-deploy-on-push")
		source.SetNoDeployOnPush(noDeployOnPush)
	}
	if flags.Lookup("git-workdir").Changed {
		workdir, _ := flags.GetString("git-workdir")
		source.SetWorkdir(workdir)
	}
	return h.setGitSourceBuilder(flags, source)
}

// Parse --git-builder and --git-* flags
func (h *ServiceHandler) setGitSourceBuilder(flags *pflag.FlagSet, source *koyeb.GitSource) (*koyeb.GitSource, error) {
	builder, _ := flags.GetString("git-builder")
	if builder != "buildpack" && builder != "docker" {
		return nil, &errors.CLIError{
			What: "Error while updating the service",
			Why:  "the --git-builder is invalid",
			Additional: []string{
				"The --git-builder must be either 'buildpack' or 'docker'",
			},
			Orig:     nil,
			Solution: "Fix the --git-builder and try again",
		}
	}
	// If docker builder arguments are specified with --git-builder=buildpack,
	// or if --git-builder is not specified but the current source is a docker
	// builder, return an error
	if flags.Lookup("git-docker-dockerfile").Changed ||
		flags.Lookup("git-docker-entrypoint").Changed ||
		flags.Lookup("git-docker-command").Changed ||
		flags.Lookup("git-docker-args").Changed ||
		flags.Lookup("git-docker-target").Changed ||
		(flags.Lookup("git-builder").Changed && builder == "docker") ||
		source.HasDocker() {

		// If --git-builder has not been provided, but the current source is a buildpack builder.
		if !flags.Lookup("git-builder").Changed && source.HasBuildpack() {
			return nil, &errors.CLIError{
				What: "Error while updating the service",
				Why:  "invalid flag combination",
				Additional: []string{
					"The arguments --git-docker-* are used to configure the docker builder, but the current builder is a buildpack builder",
				},
				Orig:     nil,
				Solution: "Add --git-builder=docker to the arguments to configure the docker builder",
			}
		}
		builder, err := h.parseGitSourceDockerBuilder(flags, source.GetDocker())
		if err != nil {
			return nil, err
		}
		source.Buildpack = nil
		source.SetDocker(*builder)
	}
	if flags.Lookup("git-buildpack-build-command").Changed ||
		flags.Lookup("git-build-command").Changed ||
		flags.Lookup("git-buildpack-run-command").Changed ||
		flags.Lookup("git-run-command").Changed ||
		(flags.Lookup("git-builder").Changed && builder == "buildpack") ||
		source.HasBuildpack() {

		// If --git-builder has not been provided, but the current source is a buildpack builder.
		if !flags.Lookup("git-builder").Changed && source.HasDocker() {
			return nil, &errors.CLIError{
				What: "Error while updating the service",
				Why:  "invalid flag combination",
				Additional: []string{
					"The arguments --git-buildpack-* are used to configure the buildpack builder, but the current builder is a docker builder",
				},
				Orig:     nil,
				Solution: "Add --git-builder=buildpack to the arguments to configure the buildpack builder",
			}
		}

		builder, err := h.parseGitSourceBuildpackBuilder(flags, source)
		if err != nil {
			return nil, err
		}
		source.SetBuildpack(*builder)
		source.Docker = nil
	}
	return source, nil
}

// Parse --git-buildpack-* flags
func (h *ServiceHandler) parseGitSourceBuildpackBuilder(flags *pflag.FlagSet, source *koyeb.GitSource) (*koyeb.BuildpackBuilder, error) {
	builder := source.GetBuildpack()
	// Legacy options for backward compatibility. We prefer
	// --git-buildpack-build-command and --git-buildpack-run-command over --git-build-command and --git-run-command
	buildCommand, _ := flags.GetString("git-build-command")
	buildpackBuildCommand, _ := flags.GetString("git-buildpack-build-command")
	runCommand, _ := flags.GetString("git-run-command")
	buildpackRunCommand, _ := flags.GetString("git-buildpack-run-command")

	if flags.Lookup("git-build-command").Changed && flags.Lookup("git-buildpack-build-command").Changed {
		return nil, &errors.CLIError{
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
	if flags.Lookup("git-run-command").Changed && flags.Lookup("git-buildpack-run-command").Changed {
		return nil, &errors.CLIError{
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
	if flags.Lookup("git-build-command").Changed {
		source.SetBuildCommand(buildCommand)
		builder.SetBuildCommand(buildCommand)
	} else if flags.Lookup("git-buildpack-build-command").Changed {
		source.SetBuildCommand(buildpackBuildCommand)
		builder.SetBuildCommand(buildpackBuildCommand)
	}
	if flags.Lookup("git-run-command").Changed {
		source.SetRunCommand(runCommand)
		builder.SetRunCommand(runCommand)
	} else if flags.Lookup("git-buildpack-run-command").Changed {
		source.SetRunCommand(buildpackRunCommand)
		builder.SetRunCommand(buildpackRunCommand)
	}
	if flags.Lookup("privileged").Changed {
		privileged, _ := flags.GetBool("privileged")
		builder.SetPrivileged(privileged)
	}
	return &builder, nil
}

// Parse --git-docker-* flags
func (h *ServiceHandler) parseGitSourceDockerBuilder(flags *pflag.FlagSet, builder koyeb.DockerBuilder) (*koyeb.DockerBuilder, error) {
	if flags.Lookup("git-docker-dockerfile").Changed {
		dockerfile, _ := flags.GetString("git-docker-dockerfile")
		builder.SetDockerfile(dockerfile)
	}
	if flags.Lookup("git-docker-entrypoint").Changed {
		entrypoint, _ := flags.GetStringSlice("git-docker-entrypoint")
		builder.SetEntrypoint(entrypoint)
	}
	if flags.Lookup("git-docker-command").Changed {
		command, _ := flags.GetString("git-docker-command")
		builder.SetCommand(command)
	}
	if flags.Lookup("git-docker-args").Changed {
		args, _ := flags.GetStringSlice("git-docker-args")
		builder.SetArgs(args)
	}
	if flags.Lookup("git-docker-target").Changed {
		target, _ := flags.GetString("git-docker-target")
		builder.SetTarget(target)
	}
	if flags.Lookup("privileged").Changed {
		privileged, _ := flags.GetBool("privileged")
		builder.SetPrivileged(privileged)
	}
	return &builder, nil
}

// Parse --archive-* flags
func (h *ServiceHandler) parseArchiveSource(flags *pflag.FlagSet, source *koyeb.ArchiveSource) (*koyeb.ArchiveSource, error) {
	if flags.Lookup("archive").Changed {
		archive, _ := flags.GetString("archive")
		source.SetId(archive)
	}
	return h.setArchiveSourceBuilder(flags, source)
}

// Parse --archive-builder and --archive-* flags
func (h *ServiceHandler) setArchiveSourceBuilder(flags *pflag.FlagSet, source *koyeb.ArchiveSource) (*koyeb.ArchiveSource, error) {
	builder, _ := flags.GetString("archive-builder")
	if builder != "buildpack" && builder != "docker" {
		return nil, &errors.CLIError{
			What: "Error while updating the service",
			Why:  "the --archive-builder is invalid",
			Additional: []string{
				"The --archive-builder must be either 'buildpack' or 'docker'",
			},
			Orig:     nil,
			Solution: "Fix the --archive-builder and try again",
		}
	}
	// If docker builder arguments are specified with --archive-builder=buildpack,
	// or if --archive-builder is not specified but the current source is a docker
	// builder, return an error
	if flags.Lookup("archive-docker-dockerfile").Changed ||
		flags.Lookup("archive-docker-entrypoint").Changed ||
		flags.Lookup("archive-docker-command").Changed ||
		flags.Lookup("archive-docker-args").Changed ||
		flags.Lookup("archive-docker-target").Changed ||
		(flags.Lookup("archive-builder").Changed && builder == "docker") ||
		source.HasDocker() {

		// If --archive-builder has not been provided, but the current source is a buildpack builder.
		if !flags.Lookup("archive-builder").Changed && source.HasBuildpack() {
			return nil, &errors.CLIError{
				What: "Error while updating the service",
				Why:  "invalid flag combination",
				Additional: []string{
					"The arguments --archive-docker-* are used to configure the docker builder, but the current builder is a buildpack builder",
				},
				Orig:     nil,
				Solution: "Add --archive-builder=docker to the arguments to configure the docker builder",
			}
		}
		builder, err := h.parseArchiveSourceDockerBuilder(flags, source.GetDocker())
		if err != nil {
			return nil, err
		}
		source.Buildpack = nil
		source.SetDocker(*builder)
	}
	if flags.Lookup("archive-buildpack-build-command").Changed ||
		flags.Lookup("archive-buildpack-run-command").Changed ||
		(flags.Lookup("archive-builder").Changed && builder == "buildpack") ||
		source.HasBuildpack() {

		// If --archive-builder has not been provided, but the current source is a buildpack builder.
		if !flags.Lookup("archive-builder").Changed && source.HasDocker() {
			return nil, &errors.CLIError{
				What: "Error while updating the service",
				Why:  "invalid flag combination",
				Additional: []string{
					"The arguments --archive-buildpack-* are used to configure the buildpack builder, but the current builder is a docker builder",
				},
				Orig:     nil,
				Solution: "Add --archive-builder=buildpack to the arguments to configure the buildpack builder",
			}
		}

		builder, err := h.parseArchiveSourceBuildpackBuilder(flags, source)
		if err != nil {
			return nil, err
		}
		source.SetBuildpack(*builder)
		source.Docker = nil
	}
	return source, nil
}

// Parse --archive-buildpack-* flags
func (h *ServiceHandler) parseArchiveSourceBuildpackBuilder(flags *pflag.FlagSet, source *koyeb.ArchiveSource) (*koyeb.BuildpackBuilder, error) {
	builder := source.GetBuildpack()
	buildpackBuildCommand, _ := flags.GetString("archive-buildpack-build-command")
	buildpackRunCommand, _ := flags.GetString("archive-buildpack-run-command")

	if flags.Lookup("archive-buildpack-build-command").Changed {
		builder.SetBuildCommand(buildpackBuildCommand)
	}
	if flags.Lookup("archive-buildpack-run-command").Changed {
		builder.SetRunCommand(buildpackRunCommand)
	}
	if flags.Lookup("privileged").Changed {
		privileged, _ := flags.GetBool("privileged")
		builder.SetPrivileged(privileged)
	}
	return &builder, nil
}

// Parse --archive-docker-* flags
func (h *ServiceHandler) parseArchiveSourceDockerBuilder(flags *pflag.FlagSet, builder koyeb.DockerBuilder) (*koyeb.DockerBuilder, error) {
	if flags.Lookup("archive-docker-dockerfile").Changed {
		dockerfile, _ := flags.GetString("archive-docker-dockerfile")
		builder.SetDockerfile(dockerfile)
	}
	if flags.Lookup("archive-docker-entrypoint").Changed {
		entrypoint, _ := flags.GetStringSlice("archive-docker-entrypoint")
		builder.SetEntrypoint(entrypoint)
	}
	if flags.Lookup("archive-docker-command").Changed {
		command, _ := flags.GetString("archive-docker-command")
		builder.SetCommand(command)
	}
	if flags.Lookup("archive-docker-args").Changed {
		args, _ := flags.GetStringSlice("archive-docker-args")
		builder.SetArgs(args)
	}
	if flags.Lookup("archive-docker-target").Changed {
		target, _ := flags.GetString("archive-docker-target")
		builder.SetTarget(target)
	}
	if flags.Lookup("privileged").Changed {
		privileged, _ := flags.GetBool("privileged")
		builder.SetPrivileged(privileged)
	}
	return &builder, nil
}

func (h *ServiceHandler) ParseArchiveIgnoreDirectories(flags *pflag.FlagSet, handler *ArchiveHandler) error {
	ignoreDirectories, err := flags.GetStringSlice("archive-ignore-dir")
	if err != nil {
		return err
	}
	handler.ParseIgnoreDirectories(ignoreDirectories)
	return nil

}

// DeploymentDefinition contains the keys "env", "scalings" and "instance_types"
// which are lists of objects with the key "scopes" containing the names of the
// regions where the ressource should be deployed, such as:
//
//	"definition": {
//		"env": [{
//			"key": "DATABASE_URL",
//			"scopes": ["region:fra"],
//			"secret": "<secret value>"
//		}],
//		"scalings": [{
//			"max": 1,
//			"min": 1,
//			"scopes": ["region:fra"]
//		}],
//		"instance_types": [{
//			"scopes": ["region:fra"],
//			"type": "nano"
//		}],
//	}
//
// setRegions updates these list of "scopes":
// - removes the scope that are not in the list of regions (if region has been removed)
// - add the scope that are in the list of regions but not in the list of scopes (if region has been added)
//
// For now, this is dumb as we do not have a feature fine grained update of what
// is exposed for a given region. For example, while the API allows to update an
// environment variable to have a specific value for a given region and another
// value for another region, the CLI and the console do not allow to do that.
func (h *ServiceHandler) setRegions(definition *koyeb.DeploymentDefinition, regions []string) {
	definition.SetRegions(regions)

	updateScopes := func(regions []string, currentScopes []string) []string {
		regionsMap := make(map[string]bool)
		for _, region := range regions {
			regionsMap[region] = false
		}

		newScopes := []string{}
		for _, scope := range currentScopes {
			// Append scope if it is not a region scope (even if we don't support other scopes for now)
			if !strings.HasPrefix(scope, "region:") {
				newScopes = append(newScopes, scope)
			} else {
				region := strings.TrimPrefix(scope, "region:")
				if _, exists := regionsMap[region]; exists {
					newScopes = append(newScopes, scope)
					regionsMap[region] = true
				}
			}
		}

		// Add new regions
		for region, exists := range regionsMap {
			if !exists {
				newScopes = append(newScopes, fmt.Sprintf("region:%s", region))
			}
		}
		return newScopes
	}

	for idx, env := range definition.Env {
		definition.Env[idx].Scopes = updateScopes(regions, env.Scopes)
	}
	for idx, scalings := range definition.Scalings {
		definition.Scalings[idx].Scopes = updateScopes(regions, scalings.Scopes)
	}
	for idx, instanceTypes := range definition.InstanceTypes {
		definition.InstanceTypes[idx].Scopes = updateScopes(regions, instanceTypes.Scopes)
	}
}

// The service name must be in the form <app>/<service> or <service>.
var ServiceNameRegexp = regexp.MustCompile(`^(?:(?P<app_name>[^/]+)/)?(?P<service_name>[^/]+)$`)

// parseServiceName returns the service name in the form <app>/<service>, or
// <service> if the application name is not specified.
//
// For context, all the "koyeb service" commands accept two syntaxes to specify
// the service name:
// - <app>/<service>
// - <service> in which case the application can be specified with the --app flag
//
// This function returns the service name in the form <app>/<service> or
// <service>, and returns an error if the service name is invalid, or if both
// --app and <app>/<service> are specified but the application names do not
// match.
func (h *ServiceHandler) parseServiceName(cmd *cobra.Command, serviceName string) (string, error) {
	match := ServiceNameRegexp.FindStringSubmatch(serviceName)

	if match == nil {
		return "", &errors.CLIError{
			What:       "Invalid service name",
			Why:        "the service name must be in the form <app>/<service> or <service>",
			Additional: nil,
			Orig:       nil,
			Solution:   "Fix the service name and try again",
		}
	}

	appFlagValue, _ := cmd.Flags().GetString("app")
	if appFlagValue != "" {
		if match[ServiceNameRegexp.SubexpIndex("app_name")] != "" && appFlagValue != match[ServiceNameRegexp.SubexpIndex("app_name")] {
			return "", &errors.CLIError{
				What:       "Inconsistent values for the --app flag and the service name",
				Why:        "the application name provided with the --app flag and the application name in <app>/<service> do not match",
				Additional: nil,
				Orig:       nil,
				Solution:   "Update/remove the --app flag, or the application in the service name, and try again",
			}
		}
		return fmt.Sprintf(
			"%s/%s",
			appFlagValue,
			match[ServiceNameRegexp.SubexpIndex("service_name")],
		), nil
	}
	if match[ServiceNameRegexp.SubexpIndex("app_name")] == "" {
		return match[ServiceNameRegexp.SubexpIndex("service_name")], nil
	}
	return fmt.Sprintf(
		"%s/%s",
		match[ServiceNameRegexp.SubexpIndex("app_name")],
		match[ServiceNameRegexp.SubexpIndex("service_name")],
	), nil
}

// parseServiceNameWithoutApp is similar to parseServiceName, but does not return the application name.
func (h *ServiceHandler) parseServiceNameWithoutApp(cmd *cobra.Command, serviceName string) (string, error) {
	name, err := h.parseServiceName(cmd, serviceName)
	if err != nil {
		return "", err
	}
	split := strings.SplitN(name, "/", 2)
	if len(split) == 1 {
		return split[0], nil
	}
	return split[1], nil
}

// parseAppName is similar to parseServiceName, but returns the application name. If the application name is not specified, an error is returned.
func (h *ServiceHandler) parseAppName(cmd *cobra.Command, serviceName string) (string, error) {
	name, err := h.parseServiceName(cmd, serviceName)
	if err != nil {
		return "", err
	}
	split := strings.SplitN(name, "/", 2)
	if len(split) == 1 {
		return "", &errors.CLIError{
			What:       "Missing application name",
			Why:        "the application name has not been provided",
			Additional: nil,
			Orig:       nil,
			Solution:   "Set the flag --app, or specify the application name in the service name with the form <app>/<service>",
		}
	}
	return split[0], nil
}

// checkDockerImage calls the API /v1/docker-helper/verify to check the validity
// of the docker image in the given source. It returns nil if the image is
// valid, or an error if the image is invalid.
func (h *ServiceHandler) checkDockerImage(ctx *CLIContext, source *koyeb.DockerSource) error {
	req := ctx.Client.DockerHelperApi.VerifyDockerImage(ctx.Context).Image(source.GetImage())

	secret := source.GetImageRegistrySecret()
	if len(secret) > 0 {
		secretID, err := ResolveSecretArgs(ctx, secret)
		if err != nil {
			return &errors.CLIError{
				What:       fmt.Sprintf("Error while checking the validity of the docker image `%s`", source.GetImage()),
				Why:        "the secret provided with the --docker-private-registry-secret flag is invalid",
				Additional: []string{},
				Orig:       nil,
				Solution:   "Create a secret with `koyeb secret create --type registry-<type>` and provide the secret name with the --docker-private-registry-secret flag",
			}
		}
		req = req.SecretId(secretID)
	}

	res, resp, err := req.Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while checking the validity of the docker image `%s`", source.GetImage()),
			err,
			resp,
		)
	}

	if !res.GetSuccess() {
		return &errors.CLIError{
			What: "Error while checking the validity of the docker image",
			Why:  res.GetReason(),
			Additional: []string{
				fmt.Sprintf("Make sure the image name `%s` is correct and that the image exists.", source.GetImage()),
				"If the image requires authentication, make sure to provide the parameter --docker-private-registry-secret.",
			},
			Orig:     nil,
			Solution: "Fix the image name or provide the required authentication and try again",
		}
	}
	return nil
}

// Parse --volumes
func (h *ServiceHandler) parseVolumes(ctx *CLIContext, flags *pflag.FlagSet, currentVolumes []koyeb.DeploymentVolume) ([]koyeb.DeploymentVolume, error) {
	wrappedResolveVolumeId := func(value string) (string, error) {
		return h.ResolveVolumeArgs(ctx, value)
	}

	return parseListFlags("volumes", flags_list.GetNewVolumeListFromFlags(wrappedResolveVolumeId), flags, currentVolumes)
}

// Parse --config-file
func (h *ServiceHandler) parseConfigFiles(ctx *CLIContext, flags *pflag.FlagSet, currentFiles []koyeb.ConfigFile) ([]koyeb.ConfigFile, error) {
	return parseListFlags("config-file", flags_list.GetNewConfigFilestListFromFlags(), flags, currentFiles)
}

// DeploymentStrategy is a type alias for koyeb.DeploymentStrategyType which implements the pflag.Value interface.
type DeploymentStrategy koyeb.DeploymentStrategyType

func (s *DeploymentStrategy) String() string {
	return string(*s)
}

func (s *DeploymentStrategy) Set(value string) error {
	switch value {
	case "rolling":
		*s = DeploymentStrategy(koyeb.DEPLOYMENTSTRATEGYTYPE_ROLLING)
	case "blue-green":
		*s = DeploymentStrategy(koyeb.DEPLOYMENTSTRATEGYTYPE_BLUE_GREEN)
	case "immediate":
		*s = DeploymentStrategy(koyeb.DEPLOYMENTSTRATEGYTYPE_IMMEDIATE)
	default:
		return fmt.Errorf("invalid deployment strategy: %s. Valid values are: rolling, blue-green, immediate.", value)
	}
	return nil
}

func (s *DeploymentStrategy) Type() string {
	return "STRATEGY"
}
