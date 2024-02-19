package koyeb

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
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

# Build and deploy a GitHub repository using buildpack (default), set the environment variable PORT, and expose the port 9000 to the root route
$> koyeb service create myservice --app myapp --git github.com/koyeb/example-flask --git-branch main --env PORT=9000 --port 9000:http --route /:9000

# Build and deploy a GitHub repository using docker
$> koyeb service create myservice --app myapp --git github.com/org/name --git-branch main --git-builder docker`,
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			createService := koyeb.NewCreateServiceWithDefaults()
			createDefinition := koyeb.NewDeploymentDefinitionWithDefaults()

			err := parseServiceDefinitionFlags(ctx, cmd.Flags(), createDefinition)
			if err != nil {
				return err
			}

			serviceName, err := parseServiceNameWithoutApp(cmd, args[0])
			if err != nil {
				return err
			}

			createDefinition.Name = koyeb.PtrString(serviceName)
			createService.SetDefinition(*createDefinition)
			return h.Create(ctx, cmd, args, createService)
		}),
	}
	addServiceDefinitionFlags(createServiceCmd.Flags())
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

	logsServiceCmd := &cobra.Command{
		Use:     "logs NAME",
		Aliases: []string{"l", "log"},
		Short:   "Get the service logs",
		Args:    cobra.ExactArgs(1),
		RunE:    WithCLIContext(h.Logs),
	}
	logsServiceCmd.Flags().StringP("app", "a", "", "Service application")
	logsServiceCmd.Flags().String("instance", "", "Instance")
	logsServiceCmd.Flags().StringP("type", "t", "", "Type (runtime, build)")
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
$> koyeb service update myapp/myservice --env PORT=8001 --env '!DEBUG'`,
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			serviceName, err := parseServiceName(cmd, args[0])
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

			// zero the sha to make sure that the latest sha is fetched
			if updateDef.Git != nil {
				updateDef.Git.Sha = koyeb.PtrString("")
			}

			err = parseServiceDefinitionFlags(ctx, cmd.Flags(), updateDef)
			if err != nil {
				return err
			}
			updateService.SetDefinition(*updateDef)
			return h.Update(ctx, cmd, args, updateService)
		}),
	}
	addServiceDefinitionFlags(updateServiceCmd.Flags())
	updateServiceCmd.Flags().StringP("app", "a", "", "Service application")
	updateServiceCmd.Flags().Bool("override", false, "Override the service configuration with the new configuration instead of merging them")
	serviceCmd.AddCommand(updateServiceCmd)

	redeployServiceCmd := &cobra.Command{
		Use:   "redeploy NAME",
		Short: "Redeploy service",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.ReDeploy),
	}
	redeployServiceCmd.Flags().StringP("app", "a", "", "Service application")
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

func addServiceDefinitionFlags(flags *pflag.FlagSet) {
	// Global flags
	flags.String("type", "web", `Service type, either "web" or "worker"`)

	flags.StringSlice(
		"regions",
		[]string{},
		"Add a region where the service is deployed. You can specify this flag multiple times to deploy the service in multiple regions.\n"+
			"To update a service and remove a region, prefix the region name with '!', for example --region '!par'\n"+
			"If the region is not specified on service creation, the service is deployed in fra\n",
	)
	flags.StringSlice(
		"env",
		[]string{},
		"Update service environment variables using the format KEY=VALUE, for example --env FOO=bar\n"+
			"To use the value of a secret as an environment variable, specify the secret name preceded by @, for example --env FOO=@bar\n"+
			"To delete an environment variable, prefix its name with '!', for example --env '!FOO'\n",
	)
	flags.String("instance-type", "nano", "Instance type")
	flags.Int64("scale", 1, "Set both min-scale and max-scale")
	flags.Int64("min-scale", 1, "Min scale")
	flags.Int64("max-scale", 1, "Max scale")
	flags.Int64("autoscaling-average-cpu", 0, "Target CPU usage (in %) to trigger a scaling event. Set to 0 to disable CPU autoscaling.")
	flags.Int64("autoscaling-average-mem", 0, "Target memory usage (in %) to trigger a scaling event. Set to 0 to disable memory autoscaling.")
	flags.Int64("autoscaling-requests-per-second", 0, "Target requests per second to trigger a scaling event. Set to 0 to disable requests per second autoscaling.")
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
			"PROTOCOL defaults to \"http\". Supported protocols are \"http\" and \"http2\"\n"+
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

	// Git service
	flags.String("git", "", "Git repository")
	flags.String("git-branch", "main", "Git branch")
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
	flags.String("git-docker-command", "", "Docker CMD")
	flags.StringSlice("git-docker-args", []string{}, "Arguments for the Docker CMD")
	flags.String("git-docker-target", "", "Docker target")

	// Docker service
	flags.String("docker", "", "Docker image")
	flags.String("docker-private-registry-secret", "", "Docker private registry secret")
	flags.StringSlice("docker-entrypoint", []string{}, "Docker entrypoint")
	flags.String("docker-command", "", "Docker command")
	flags.StringSlice("docker-args", []string{}, "Docker args")

	// Configure aliases: for example, allow user to use --port instead of --ports
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

func parseServiceDefinitionFlags(ctx *CLIContext, flags *pflag.FlagSet, definition *koyeb.DeploymentDefinition) error {
	type_, err := parseType(flags, definition.GetType())
	if err != nil {
		return err
	}
	definition.SetType(type_)

	skipCache, _ := flags.GetBool("skip-cache")
	definition.SetSkipCache(skipCache)

	envs, err := parseEnv(flags, definition.Env)
	if err != nil {
		return err
	}
	definition.SetEnv(envs)

	definition.SetInstanceTypes(parseInstanceType(flags, definition.GetInstanceTypes()))

	ports, err := parsePorts(definition.GetType(), flags, definition.Ports)
	if err != nil {
		return err
	}
	definition.SetPorts(ports)

	routes, err := parseRoutes(definition.GetType(), flags, definition.Routes)
	if err != nil {
		return err
	}
	definition.SetRoutes(routes)

	if definition.GetType() == koyeb.DEPLOYMENTDEFINITIONTYPE_WEB {
		err = setDefaultPortsAndRoutes(definition, definition.Ports, definition.Routes)
		if err != nil {
			return err
		}
	}

	definition.SetScalings(parseScalings(flags, definition.Scalings))

	healthchecks, err := parseChecks(definition.GetType(), flags, definition.HealthChecks)
	if err != nil {
		return err
	}
	definition.SetHealthChecks(healthchecks)

	regions, err := parseRegions(flags, definition.GetRegions())
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
	setRegions(definition, regions)

	err = setSource(ctx, definition, flags)
	if err != nil {
		return err
	}
	return nil
}

// Parse --type
func parseType(flags *pflag.FlagSet, currentType koyeb.DeploymentDefinitionType) (koyeb.DeploymentDefinitionType, error) {
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
func parseInstanceType(flags *pflag.FlagSet, currentInstanceTypes []koyeb.DeploymentInstanceType) []koyeb.DeploymentInstanceType {
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

// parseListFlags is the generic function parsing --env, --port, --routes, --checks and --regions.
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
	newItems := flags_list.ParseListFlags[T](listFlags, currentItems)
	return newItems, nil
}

// Parse --env
func parseEnv(flags *pflag.FlagSet, currentEnv []koyeb.DeploymentEnv) ([]koyeb.DeploymentEnv, error) {
	return parseListFlags("env", flags_list.NewEnvListFromFlags, flags, currentEnv)
}

// Parse --ports
func parsePorts(type_ koyeb.DeploymentDefinitionType, flags *pflag.FlagSet, currentPorts []koyeb.DeploymentPort) ([]koyeb.DeploymentPort, error) {
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
func parseRoutes(type_ koyeb.DeploymentDefinitionType, flags *pflag.FlagSet, currentRoutes []koyeb.DeploymentRoute) ([]koyeb.DeploymentRoute, error) {
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
func getDeploymentPort(portNumber int64) []koyeb.DeploymentPort {
	newPort := koyeb.NewDeploymentPortWithDefaults()
	newPort.SetPort(portNumber)
	newPort.SetProtocol("http")
	return []koyeb.DeploymentPort{*newPort}
}

// Set default route to `portNumber` and `/`
func getDeploymentRoute(portNumber int64) []koyeb.DeploymentRoute {
	newRoute := koyeb.NewDeploymentRouteWithDefaults()
	newRoute.SetPath("/")
	newRoute.SetPort(portNumber)
	return []koyeb.DeploymentRoute{*newRoute}
}

// Dynamically sets the defaults ports and routes for "web" services
func setDefaultPortsAndRoutes(definition *koyeb.DeploymentDefinition, currentPorts []koyeb.DeploymentPort, currentRoutes []koyeb.DeploymentRoute) error {
	switch {
	// If no route and no port is specified, add the default route and port
	case len(currentPorts) == 0 && len(currentRoutes) == 0:
		definition.SetPorts(getDeploymentPort(8000))
		definition.SetRoutes(getDeploymentRoute(8000))

	// One or more port set, no route set
	case len(currentPorts) > 0 && len(currentRoutes) == 0:
		// Two or more ports are set
		if len(currentPorts) > 1 {
			return &errors.CLIError{
				What: "Error while configuring the service",
				Why:  `your service has two or more ports set but no matching routes`,
				Additional: []string{
					"For each port, you must specify a matching route with the --routes flag",
				},
				Orig:     nil,
				Solution: "Set the routes and try again",
			}
		}
		// One port set
		portNumber := currentPorts[0].GetPort()
		definition.SetRoutes(getDeploymentRoute(portNumber))

	// One or more route set, no port set
	case len(currentRoutes) > 0 && len(currentPorts) == 0:
		// Two or more routes are set
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
		// One route set
		portNumber := currentRoutes[0].GetPort()
		definition.SetPorts(getDeploymentPort(portNumber))
	}
	return nil
}

// Parse --checks
func parseChecks(type_ koyeb.DeploymentDefinitionType, flags *pflag.FlagSet, currentHealthChecks []koyeb.DeploymentHealthCheck) ([]koyeb.DeploymentHealthCheck, error) {
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
	return newChecks, nil
}

// Parse --regions
func parseRegions(flags *pflag.FlagSet, currentRegions []string) ([]string, error) {
	newRegions, err := parseListFlags("regions", flags_list.NewRegionsListFromFlags, flags, currentRegions)
	if err != nil {
		return nil, err
	}
	// For new services, if no region is specified, add the default region
	if !flags.Lookup("regions").Changed && len(currentRegions) == 0 {
		newRegions = []string{"fra"}
	}
	return newRegions, nil
}

// Parse --min-scale and --max-scale
func parseScalings(flags *pflag.FlagSet, currentScalings []koyeb.DeploymentScaling) []koyeb.DeploymentScaling {
	var minScale int64
	var maxScale int64

	if flags.Lookup("min-scale").Changed {
		minScale, _ = flags.GetInt64("min-scale")
	} else {
		minScale, _ = flags.GetInt64("scale")
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
		setScalingsTargets(flags, scaling)
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
			setScalingsTargets(flags, &currentScalings[idx])
		}
	}
	return currentScalings
}

// setScalingsTargets updates the scaling targets in a koyeb.DeploymentScaling object based on the specified flags.
// It checks for changes in flags related to autoscaling parameters (CPU, memory, and requests per second).
// If a change is detected in any of these flags, the function either updates an existing scaling target or creates a new one.
//
// setScalingsTargets performs the following steps:
// 1. For each autoscaling parameter (CPU, memory, requests per second), it checks if the corresponding flag has changed.
// 2. If a flag has changed, it iterates over the existing targets to find a matching target type.
// 3. If a matching target is found, it updates the target with the new value or removes the target if the value is 0.
// 4. If no matching target is found, it creates a new target with the specified value and appends it to the scaling targets.
//
// Note: there is no way to easily avoid the code duplication in this function
// because NewDeploymentScalingTarget{AverageCPU,AverageMem,RequestsPerSecond}
// do not implement a common interface.
func setScalingsTargets(flags *pflag.FlagSet, scaling *koyeb.DeploymentScaling) {
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
}

// Parse --git-* and --docker-* flags to set deployment.Git or deployment.Docker
func setSource(ctx *CLIContext, definition *koyeb.DeploymentDefinition, flags *pflag.FlagSet) error {
	hasGitFlags := false
	hasDockerFlags := false
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Changed {
			if strings.HasPrefix(flag.Name, "git") {
				hasGitFlags = true
			} else if strings.HasPrefix(flag.Name, "docker") {
				hasDockerFlags = true
			}
		}
	})
	if hasGitFlags && hasDockerFlags {
		return &errors.CLIError{
			What: "Error while updating the service",
			Why:  "invalid flag combination",
			Additional: []string{
				"Your service has both --git-* and --docker-* flags, which is not allowed.",
				"To build a GitHub repository, specify --git-*",
				"To deploy a Docker image hosted on the Docker Hub or a private registry, specify --docker-*",
			},
			Orig:     nil,
			Solution: "Fix the flags and try again",
		}
	}
	if hasDockerFlags || definition.HasDocker() {
		docker := definition.GetDocker()
		source, err := parseDockerSource(ctx, flags, &docker)
		if err != nil {
			return err
		}
		definition.SetDocker(*source)
		definition.Git = nil
	} else if hasGitFlags || definition.HasGit() {
		git := definition.GetGit()
		source, err := parseGitSource(flags, &git)
		if err != nil {
			return err
		}
		definition.Docker = nil
		definition.SetGit(*source)
	}
	return nil
}

// Parse --docker-* flags
func parseDockerSource(ctx *CLIContext, flags *pflag.FlagSet, source *koyeb.DockerSource) (*koyeb.DockerSource, error) {
	// docker-private-registry-secret needs to be parsed first, because checkDockerImage reads it
	if flags.Lookup("docker-private-registry-secret").Changed {
		secret, _ := flags.GetString("docker-private-registry-secret")
		source.SetImageRegistrySecret(secret)
	}
	if flags.Lookup("docker").Changed {
		image, _ := flags.GetString("docker")
		source.SetImage(image)
		if err := checkDockerImage(ctx, source); err != nil {
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
func parseGitSource(flags *pflag.FlagSet, source *koyeb.GitSource) (*koyeb.GitSource, error) {
	if flags.Lookup("git").Changed {
		repository, _ := flags.GetString("git")
		source.SetRepository(repository)
	}
	if source.GetBranch() == "" || flags.Lookup("git-branch").Changed {
		branch, _ := flags.GetString("git-branch")
		source.SetBranch(branch)
	}
	if flags.Lookup("git-no-deploy-on-push").Changed {
		noDeployOnPush, _ := flags.GetBool("git-no-deploy-on-push")
		source.SetNoDeployOnPush(noDeployOnPush)
	}
	if flags.Lookup("git-workdir").Changed {
		workdir, _ := flags.GetString("git-workdir")
		source.SetWorkdir(workdir)
	}
	return setGitSourceBuilder(flags, source)
}

// Parse --git-builder and --git-docker-*
func setGitSourceBuilder(flags *pflag.FlagSet, source *koyeb.GitSource) (*koyeb.GitSource, error) {
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
		builder, err := parseGitSourceDockerBuilder(flags, source.GetDocker())
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

		builder, err := parseGitSourceBuildpackBuilder(flags, source)
		if err != nil {
			return nil, err
		}
		source.SetBuildpack(*builder)
		source.Docker = nil
	}
	return source, nil
}

// Parse --git-buildpack-* flags
func parseGitSourceBuildpackBuilder(flags *pflag.FlagSet, source *koyeb.GitSource) (*koyeb.BuildpackBuilder, error) {
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
func parseGitSourceDockerBuilder(flags *pflag.FlagSet, builder koyeb.DockerBuilder) (*koyeb.DockerBuilder, error) {
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
func setRegions(definition *koyeb.DeploymentDefinition, regions []string) {
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
func parseServiceName(cmd *cobra.Command, serviceName string) (string, error) {
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
				What:       "Inconsitent values for the --app flag and the service name",
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
func parseServiceNameWithoutApp(cmd *cobra.Command, serviceName string) (string, error) {
	name, err := parseServiceName(cmd, serviceName)
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
func parseAppName(cmd *cobra.Command, serviceName string) (string, error) {
	name, err := parseServiceName(cmd, serviceName)
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
func checkDockerImage(ctx *CLIContext, source *koyeb.DockerSource) error {
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
			What: fmt.Sprintf("Error while checking the validity of the docker image `%s`", source.GetImage()),
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
