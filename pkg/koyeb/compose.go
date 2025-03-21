package koyeb

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/ghodss/yaml"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

func NewComposeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compose KOYEB_COMPOSE_FILE_PATH",
		Short: "Create koyeb resources read from koyeb-compose.yaml file",
		Example: `
		# Init koyeb-compose.yaml file
		$> echo 'apps:
		  - name: example-app
		    services:
				- name: example-service1
				  image: nginx:latest
				  ports:
				  	- 80:80
				- name: example-service2
				  path: github.com/koyeb/golang-example-app
				  branch: main
				  ports:
				  	- 8080:8080
				  depends_on:
				  	- example-service1' > koyeb-compose.yaml
		# Apply compose file
		$> koyeb compose koyeb-compose.yaml
		`,
		Args: cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			verbose := GetBoolFlags(cmd, "verbose")

			composeFile, err := parseComposeFile(args[0])
			if err != nil {
				return err
			}

			// TODO (pawel) validate compose file
			// TODO (pawel) better error handling and tips how to fix the errors

			return NewKoyebComposeHandler().Compose(ctx, composeFile, verbose)
		}),
	}
	cmd.Flags().BoolP("verbose", "v", false, "Tails service logs to have more information about your deployment.")

	cmd.AddCommand(NewComposeLogsCmd())
	cmd.AddCommand(NewComposeDeleteCmd())

	return cmd
}

func parseComposeFile(path string) (*KoyebCompose, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	exapandEnvData := os.ExpandEnv(string(data))

	var config *KoyebCompose
	err = yaml.Unmarshal([]byte(exapandEnvData), &config)
	if err != nil {
		return nil, err
	}

	for serviceName, serviceData := range config.Services {
		// TODO (pawel) fix this hacky way of setting service name
		serviceNameCopy := strings.Clone(serviceName)
		serviceData.Name = &(serviceNameCopy)
		config.Services[serviceName] = serviceData
	}

	return config, nil
}

type KoyebCompose struct {
	koyeb.CreateApp `yaml:",inline"`
	Services        map[string]KoyebComposeService `yaml:"services"`
}

type KoyebComposeService struct {
	koyeb.DeploymentDefinition `yaml:",inline"`
	DependsOn                  []string `json:"depends_on"`
}

type KoyebComposeHandler struct{}

func NewKoyebComposeHandler() *KoyebComposeHandler {
	return &KoyebComposeHandler{}
}

func (h *KoyebComposeHandler) Compose(ctx *CLIContext, koyebCompose *KoyebCompose, verbose bool) error {
	appName := *koyebCompose.Name
	appId, err := h.CreateAppIfNotExists(ctx, appName)
	if err != nil {
		return err
	}

	servicesOrdering, err := h.OrderServices(koyebCompose.Services)
	if err != nil {
		return err
	}

	// TODO (pawel) parallelize deployments if possible (a matter of checking degree of service node in the graph)
	for _, service := range servicesOrdering {
		deploymentId, err := h.UpdateService(ctx, appId, appName, &service)
		if err != nil {
			return err
		}

		err = h.MonitorService(ctx, deploymentId, service.GetName(), verbose)
		if err != nil {
			return err
		}
	}

	log.Infof("Your app %v has been succesfully deployed 🚀", appName)

	return nil
}

func isAppMonitoringEndState(status koyeb.AppStatus) bool {
	endStates := []koyeb.AppStatus{
		koyeb.APPSTATUS_DELETED,
	}

	for _, endState := range endStates {
		if status == endState {
			return true
		}
	}
	return false
}

func (h *KoyebComposeHandler) isDeploymentMonitoringEndState(status koyeb.DeploymentStatus) bool {
	endStates := []koyeb.DeploymentStatus{
		koyeb.DEPLOYMENTSTATUS_HEALTHY,
		koyeb.DEPLOYMENTSTATUS_DEGRADED,
		koyeb.DEPLOYMENTSTATUS_UNHEALTHY,
		koyeb.DEPLOYMENTSTATUS_CANCELED,
		koyeb.DEPLOYMENTSTATUS_STOPPED,
		koyeb.DEPLOYMENTSTATUS_ERROR,
	}

	for _, endState := range endStates {
		if status == endState {
			return true
		}
	}
	return false
}

// orders services based on dependOn field (calculates topological order for services)
func (h *KoyebComposeHandler) OrderServices(services map[string]KoyebComposeService) ([]koyeb.DeploymentDefinition, error) {
	dependencies := simple.NewDirectedGraph()
	serviceNameToId := map[string]graph.Node{}
	idToServiceName := map[graph.Node]string{}
	for serviceName := range services {
		serviceNode := dependencies.NewNode()

		serviceNameToId[serviceName] = serviceNode
		idToServiceName[serviceNode] = serviceName
		dependencies.AddNode(serviceNode)
	}

	for serviceName, serviceDetails := range services {
		for _, dependency := range serviceDetails.DependsOn {
			fromNode := serviceNameToId[dependency]
			toNode, ok := serviceNameToId[serviceName]
			if !ok {
				return nil, &errors.CLIError{
					What: "failed to calculate deployment order of services",
					Why:  fmt.Sprintf("service %s depends on %s which is not defined", serviceName, dependency),
				}
			}

			edge := dependencies.NewEdge(fromNode, toNode)
			dependencies.SetEdge(edge)
		}
	}

	servicesIdOrdering, err := topo.Sort(dependencies)
	if err != nil {
		return nil, &errors.CLIError{
			What: "failed to calculate deployment order of services",
			Why:  "this probably indicates circular dependency",
		}
	}

	servicesOrdering := make([]koyeb.DeploymentDefinition, len(servicesIdOrdering))
	for i, id := range servicesIdOrdering {
		serviceName := idToServiceName[id]
		servicesOrdering[i] = services[serviceName].DeploymentDefinition
	}

	return servicesOrdering, nil
}

func (h *KoyebComposeHandler) MonitorService(ctx *CLIContext, deploymentId, serviceName string, verbose bool) error {
	var s *spinner.Spinner
	if !verbose {
		s = spinner.New(spinner.CharSets[21], 100*time.Millisecond, spinner.WithColor("green"))
		s.Start()
		defer s.Stop()
	} else {
		log.Infof("🚀 Deploying %v", serviceName)
		lq := LogsQuery{
			DeploymentId: deploymentId,
			Start:        time.Now().Format(time.RFC3339),
			Tail:         true,
			Order:        "asc",
		}
		var cancel context.CancelFunc
		ctx.Context, cancel = context.WithCancel(ctx.Context)
		defer cancel()

		go func() {
			if err := ctx.LogsClient.PrintLogs(ctx, lq); err != nil {
				log.Errorf("Error while getting logs: %s", err)
				return
			}
		}()
	}

	previousStatus := koyeb.DeploymentStatus("")
	// it's dumb as it's busy waiting but for now we don't support streaming events
	for {
		resDeployment, resp, err := ctx.Client.DeploymentsApi.GetDeployment(ctx.Context, deploymentId).Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				"Error while fetching deployment status",
				err,
				resp,
			)
		}

		currentStatus := resDeployment.Deployment.GetStatus()
		if previousStatus != currentStatus {
			previousStatus = currentStatus
			if !verbose {
				s.Suffix = fmt.Sprintf(" Deploying service %s: %s", serviceName, currentStatus)
			}
		}

		if h.isDeploymentMonitoringEndState(currentStatus) {
			break
		}

		time.Sleep(5 * time.Second)
	}

	fmt.Printf("\n")
	if previousStatus == koyeb.DEPLOYMENTSTATUS_HEALTHY {
		log.Infof("Succcessfully deployed %v ✅", serviceName)
	} else {
		log.Errorf("Failed to deploy %v deployment status: %v ❌", serviceName, previousStatus)
		return &errors.CLIError{
			What:       fmt.Sprintf("failed to deploy %v", serviceName),
			Additional: []string{"please double check koyeb compose definition"},
		}
	}

	return nil
}

// Creates app if not exists and returns app id and error if any
func (h *KoyebComposeHandler) CreateAppIfNotExists(ctx *CLIContext, appName string) (string, error) {
	appId, err := NewAppHandler().ResolveAppArgs(ctx, appName)
	if err == nil {
		return appId, nil
	}

	// if we fail to fetch the app id then failover to app creation
	createApp := koyeb.CreateApp{Name: &appName}
	resApp, resp, err := ctx.Client.AppsApi.CreateApp(ctx.Context).App(createApp).Execute()
	if err != nil {
		return "", errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while creating the application `%s`", appName),
			err,
			resp,
		)
	}

	return resApp.App.GetId(), nil
}

// Updates service or creates one if not exists
// returns deployment id
func (h *KoyebComposeHandler) UpdateService(ctx *CLIContext, appId, appName string, deploymentDefinition *koyeb.DeploymentDefinition) (string, error) {
	fullServiceName := fmt.Sprintf("%s/%s", appName, *deploymentDefinition.Name)
	serviceId, err := NewServiceHandler().ResolveServiceArgs(ctx, fullServiceName)
	if err != nil {
		// if we fail to fetch the service id then failover to service creation
		createService := &koyeb.CreateService{
			AppId:      &appId,
			Definition: deploymentDefinition,
		}

		resService, resp, err := ctx.Client.ServicesApi.CreateService(ctx.Context).Service(*createService).Execute()
		if err != nil {
			return "", errors.NewCLIErrorFromAPIError(
				"Error while creating the service",
				err,
				resp,
			)
		}

		return *resService.Service.LatestDeploymentId, nil
	}

	updateService := &koyeb.UpdateService{
		Definition: deploymentDefinition,
	}
	resService, resp, err := ctx.Client.ServicesApi.UpdateService(ctx.Context, serviceId).Service(*updateService).Execute()
	if err != nil {
		return "", errors.NewCLIErrorFromAPIError(
			"Error while creating the service",
			err,
			resp,
		)
	}

	return *resService.Service.LatestDeploymentId, nil
}
