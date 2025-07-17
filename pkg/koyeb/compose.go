package koyeb

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/ghodss/yaml"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewComposeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "compose KOYEB_COMPOSE_FILE_PATH",
		Short:   "Create Koyeb resources from a koyeb-compose.yaml file",
		Example: "koyeb compose ./examples/mesh.yaml",
		Args:    cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			verbose := GetBoolFlags(cmd, "verbose")

			composeFile, err := parseComposeFile(args[0])
			if err != nil {
				return err
			}

			return NewKoyebComposeHandler().Compose(ctx, composeFile, verbose)
		}),
	}
	cmd.Flags().BoolP("verbose", "v", false, "Tails service logs to have more information about your deployment.")

	cmd.AddCommand(NewComposeLogsCmd())
	cmd.AddCommand(NewComposeDeleteCmd())

	return cmd
}

func parseComposeFile(path string) (*koyeb.CreateCompose, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	exapandEnvData := os.ExpandEnv(string(data))

	var config *koyeb.CreateCompose
	err = yaml.Unmarshal([]byte(exapandEnvData), &config)
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

func (h *KoyebComposeHandler) Compose(ctx *CLIContext, compose *koyeb.CreateCompose, verbose bool) error {
	composeRes, _, err := ctx.Client.ComposeApi.Compose(ctx.Context).Compose(*compose).Execute()
	if err != nil {
		return err
	}

	for _, service := range composeRes.Services {
		err = h.MonitorService(ctx, *service.LatestDeploymentId, service.GetName(), verbose)
		if err != nil {
			return err
		}
	}

	log.Infof("Your app %v has been succesfully deployed 🚀", *composeRes.GetApp().Name)

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
