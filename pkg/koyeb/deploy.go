package koyeb

import (
	"strconv"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewDeployCmd() *cobra.Command {
	h := NewDeployHandler()
	appHandler := NewAppHandler()
	archiveHandler := NewArchiveHandler()
	serviceHandler := NewServiceHandler()

	deployCmd := &cobra.Command{
		Use:   "deploy <path> <app>/<service>",
		Short: "Deploy a directory to Koyeb",
		Args:  cobra.ExactArgs(2),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			if err := setProjectHeader(ctx, cmd); err != nil {
				return err
			}
			appName, err := serviceHandler.parseAppName(cmd, args[1])
			if err != nil {
				return err
			}

			appId, err := h.GetAppId(ctx, appName)
			if err != nil {
				return err
			}

			if appId != "" {
				log.Infof("Application `%s` already exists, using it", appName)
			} else {
				log.Infof("Application `%s` does not exist. Creating it", appName)
				createApp := koyeb.NewCreateAppWithDefaults()
				createApp.SetName(appName)
				createAppReply, err := appHandler.CreateApp(ctx, createApp)
				if err != nil {
					return err
				}
				appId = *createAppReply.GetApp().Id
			}

			serviceName, err := serviceHandler.parseServiceNameWithoutApp(cmd, args[1])
			if err != nil {
				return err
			}

			serviceId, err := h.GetServiceId(ctx, appId, serviceName)
			if err != nil {
				return err
			}

			err = serviceHandler.ParseArchiveIgnoreDirectories(cmd.Flags(), archiveHandler)
			if err != nil {
				return err
			}

			if serviceId == "" {
				createService := koyeb.NewCreateServiceWithDefaults()
				createDefinition := koyeb.NewDeploymentDefinitionWithDefaults()

				log.Infof("Creating and uploading an archive from `%s`", args[0])
				archiveReply, err := archiveHandler.CreateArchive(ctx, args[0])
				if err != nil {
					return err
				}

				createDefinition.Name = koyeb.PtrString(serviceName)

				archive := createDefinition.GetArchive()
				archive.Id = archiveReply.GetArchive().Id
				createDefinition.SetArchive(archive)
				createDefinition.Git = nil
				createDefinition.Docker = nil
				createService.SetDefinition(*createDefinition)

				// Update definition with the flags provided by the user.
				if err := serviceHandler.parseServiceDefinitionFlags(ctx, cmd.Flags(), createDefinition); err != nil {
					return err
				}
				if err := serviceHandler.applyCreateServiceFlags(cmd, createDefinition, createService); err != nil {
					return err
				}

				log.Infof("Creating the new service `%s`", serviceName)
				if err := serviceHandler.Create(ctx, cmd, []string{args[1]}, createService); err != nil {
					return err
				}
			} else {
				updateService := koyeb.NewUpdateServiceWithDefaults()
				updateDefinition, err := serviceHandler.latestDeploymentDefinition(ctx, serviceId, args[0], false,
					"Try again in a few seconds. If the problem persists, delete the service and create it again.")
				if err != nil {
					return err
				}

				log.Infof("Creating and uploading an archive from `%s`", args[0])
				archiveReply, err := archiveHandler.CreateArchive(ctx, args[0])
				if err != nil {
					return err
				}

				archive := updateDefinition.GetArchive()
				archive.Id = archiveReply.GetArchive().Id
				updateDefinition.SetArchive(archive)
				updateDefinition.Git = nil
				updateDefinition.Docker = nil

				// Update definition with the flags provided by the user.
				// parseServiceDefinitionFlags expects to have an archive
				// source, otherwise it would try to get the --git or --docker
				// flags which are not present.
				if err := serviceHandler.parseServiceDefinitionFlags(ctx, cmd.Flags(), updateDefinition); err != nil {
					return err
				}

				// The service account ID is immutable after creation: warn and ignore on update.
				if cmd.Flags().Lookup("service-account-id") != nil && cmd.Flags().Lookup("service-account-id").Changed {
					serviceAccountId, _ := cmd.Flags().GetString("service-account-id")
					if serviceAccountId != "" {
						log.Warnf("--service-account-id is immutable after creation, ignoring the provided value `%s`", serviceAccountId)
					}
				}

				if err := serviceHandler.applyUpdateServiceFlags(ctx, cmd, serviceId, serviceName, updateDefinition, updateService); err != nil {
					return err
				}

				log.Infof("Updating the existing service `%s`", serviceName)
				if err := serviceHandler.Update(ctx, cmd, []string{args[1]}, updateService); err != nil {
					return err
				}
			}
			return nil
		}),
	}
	deployCmd.Flags().String("app", "", "Service application. Can also be provided in the service name with the format <app>/<service>")
	deployCmd.Flags().Bool("wait", false, "Waits until the deployment is done")
	deployCmd.Flags().Duration("wait-timeout", 5*time.Minute, "Duration the wait will last until timeout")

	serviceHandler.addServiceDefinitionFlagsForAllSources(deployCmd.Flags())
	serviceHandler.addServiceDefinitionFlagsForArchiveSource(deployCmd.Flags())
	serviceHandler.addServiceAccountIdFlag(deployCmd.Flags())
	deployCmd.PersistentFlags().StringP("project", "p", "", "Workspace ID or name")
	deployCmd.PersistentFlags().String("workspace", "", "Workspace ID or name (alias for --project)")
	return deployCmd
}

func NewDeployHandler() *DeployHandler {
	return &DeployHandler{}
}

type DeployHandler struct {
}

// Return the app id if it exists, otherwise return an empty string.
func (h *DeployHandler) GetAppId(ctx *CLIContext, name string) (string, error) {
	return getAppIdByName(ctx, name)
}

// getAppIdByName returns the app id if it exists, otherwise returns an empty string.
func getAppIdByName(ctx *CLIContext, name string) (string, error) {
	page := int64(0)
	offset := int64(0)
	limit := int64(100)

	// Consume paginated results until the application is found or the end of the list is reached.
	for {
		res, resp, err := ctx.API.ListApps(ctx.Context, name, strconv.FormatInt(offset, 10), strconv.FormatInt(limit, 10))

		if err != nil {
			return "", errors.NewCLIErrorFromAPIError(
				"Error while listing applications",
				err,
				resp,
			)
		}

		for _, app := range res.GetApps() {
			if app.GetName() == name {
				return app.GetId(), nil
			}
		}

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}
	return "", nil
}

// Return the service id if it exists, otherwise return an empty string.
func (h *DeployHandler) GetServiceId(ctx *CLIContext, appId string, name string) (string, error) {
	page := int64(0)
	offset := int64(0)
	limit := int64(100)

	// Consume paginated results until the application is found or the end of the list is reached.
	for {
		res, resp, err := ctx.Client.ServicesApi.ListServices(ctx.Context).
			AppId(appId).
			Name(name).
			Offset(strconv.FormatInt(offset, 10)).
			Limit(strconv.FormatInt(limit, 10)).
			Execute()

		if err != nil {
			return "", errors.NewCLIErrorFromAPIError(
				"Error while listing services",
				err,
				resp,
			)
		}

		for _, service := range res.GetServices() {
			if service.GetName() == name {
				return service.GetId(), nil
			}
		}

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}
	return "", nil
}
