package koyeb

import (
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/briandowns/spinner"
	"github.com/ghodss/yaml"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewComposeCmd() *cobra.Command {
	return &cobra.Command{
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
			composeFile, err := parseComposeFile(args[0])
			if err != nil {
				return err
			}

			// TODO (pawel) validate compose file
			// TODO (pawel) better error handling and tips how to fix the errors

			return NewKoyebComposeHandler().Compose(ctx, composeFile)
		}),
	}
}

func parseComposeFile(path string) (*KoyebCompose, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config *KoyebCompose
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
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

func (h *KoyebComposeHandler) Compose(ctx *CLIContext, koyebCompose *KoyebCompose) error {
	appName := *koyebCompose.Name
	appId, err := h.CreateAppIfNotExists(ctx, appName)
	if err != nil {
		return err
	}

	for serviceName, serviceDetails := range koyebCompose.Services {
		serviceDefinition := &serviceDetails.DeploymentDefinition
		serviceDefinition.Name = koyeb.PtrString(serviceName)

		deploymentId, err := h.UpdateService(ctx, appId, appName, serviceDefinition)
		if err != nil {
			return err
		}

		err = h.MonitorService(ctx, deploymentId, serviceName)
		if err != nil {
			return err
		}

	}

	return nil
}

func (h *KoyebComposeHandler) isMonitoringEndState(status koyeb.DeploymentStatus) bool {
	return slices.Contains([]koyeb.DeploymentStatus{
		koyeb.DEPLOYMENTSTATUS_HEALTHY,
		koyeb.DEPLOYMENTSTATUS_DEGRADED,
		koyeb.DEPLOYMENTSTATUS_UNHEALTHY,
		koyeb.DEPLOYMENTSTATUS_CANCELED,
		koyeb.DEPLOYMENTSTATUS_STOPPED,
		koyeb.DEPLOYMENTSTATUS_ERROR,
	}, status)
}

func (h *KoyebComposeHandler) MonitorService(ctx *CLIContext, deploymentId, serviceName string) error {
	s := spinner.New(spinner.CharSets[21], 100*time.Millisecond, spinner.WithColor("green"))
	s.Start()
	defer s.Stop()

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
			s.Suffix = fmt.Sprintf(" Deploying service %s: %s", serviceName, currentStatus)
		}

		if h.isMonitoringEndState(currentStatus) {
			break
		}

		time.Sleep(5 * time.Second)
	}

	if previousStatus == koyeb.DEPLOYMENTSTATUS_HEALTHY {
		log.Infof("\nSucccessfully deployed %v ✅", serviceName)
	} else {
		log.Errorf("\nFailed to deploy %v deployment status: %v ❌", serviceName, previousStatus)
		return fmt.Errorf("failed to deploy %v", serviceName)
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
