package koyeb

import (
	"fmt"
	"strconv"

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

			// Parse the flags for the addons.
			addons, err := serviceHandler.parseAddonsFlags(ctx, cmd.Flags())
			if err != nil {
				return err
			}
			addonsHandler, err := NewAddonsHandler(addons)
			if err != nil {
				return err
			}

			// Setup the addons.
			if err := addonsHandler.Setup(ctx, args[0]); err != nil {
				return err
			}
			defer func() {
				err := addonsHandler.Cleanup(ctx)
				if err != nil {
					log.Errorf("Error while cleaning up addons: %s", err)
				}
			}()

			if serviceId == "" {
				createService := koyeb.NewCreateServiceWithDefaults()
				createDefinition := koyeb.NewDeploymentDefinitionWithDefaults()

				createDefinition.Name = koyeb.PtrString(serviceName)

				archive := createDefinition.GetArchive()
				createDefinition.SetArchive(archive)
				createDefinition.Git = nil
				createDefinition.Docker = nil
				createService.SetDefinition(*createDefinition)

				// Update definition with the flags provided by the user.
				if err := serviceHandler.parseServiceDefinitionFlags(ctx, cmd.Flags(), createDefinition); err != nil {
					return err
				}

				if err = addonsHandler.PreDeploy(ctx, createDefinition); err != nil {
					return err
				}

				log.Infof("Creating and uploading an archive from `%s`", args[0])
				archiveReply, err := archiveHandler.CreateArchive(ctx, args[0])
				if err != nil {
					return err
				}
				createDefinition.Archive.Id = archiveReply.GetArchive().Id
				createService.SetDefinition(*createDefinition)

				log.Infof("Creating the new service `%s`", serviceName)
				if err := serviceHandler.Create(ctx, cmd, []string{args[1]}, createService); err != nil {
					return err
				}
				if err = addonsHandler.PostDeploy(ctx, createDefinition); err != nil {
					return err
				}
			} else {
				updateService := koyeb.NewUpdateServiceWithDefaults()
				latestDeploy, resp, err := ctx.Client.DeploymentsApi.
					ListDeployments(ctx.Context).
					Limit("1").
					ServiceId(serviceId).
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
						Solution: "Try again in a few seconds. If the problem persists, delete the service and create it again.",
					}
				}

				updateDefinition := latestDeploy.GetDeployments()[0].Definition

				archive := updateDefinition.GetArchive()
				updateDefinition.SetArchive(archive)
				updateDefinition.Git = nil
				updateDefinition.Docker = nil
				updateService.SetDefinition(*updateDefinition)

				// Update definition with the flags provided by the user.
				// parseServiceDefinitionFlags expects to have an archive
				// source, otherwise it would try to get the --git or --docker
				// flags which are not present.
				err = serviceHandler.parseServiceDefinitionFlags(ctx, cmd.Flags(), updateDefinition)
				if err != nil {
					return err
				}

				if err = addonsHandler.PreDeploy(ctx, updateDefinition); err != nil {
					return err
				}

				log.Infof("Creating and uploading an archive from `%s`", args[0])
				archiveReply, err := archiveHandler.CreateArchive(ctx, args[0])
				if err != nil {
					return err
				}
				updateDefinition.Archive.Id = archiveReply.GetArchive().Id
				updateService.SetDefinition(*updateDefinition)

				log.Infof("Updating the existing service `%s`", serviceName)
				if err := serviceHandler.Update(ctx, cmd, []string{args[1]}, updateService); err != nil {
					return err
				}
				if err = addonsHandler.PostDeploy(ctx, updateDefinition); err != nil {
					return err
				}
			}

			return nil
		}),
	}
	deployCmd.Flags().String("app", "", "Service application. Can also be provided in the service name with the format <app>/<service>")
	serviceHandler.addServiceDefinitionFlagsForAllSources(deployCmd.Flags())
	serviceHandler.addServiceDefinitionFlagsForArchiveSource(deployCmd.Flags())

	// Add addons flags to the deploy command.
	deployCmd.Flags().StringSlice(
		"addons",
		[]string{},
		"List of addons, the addons will be executed localy before the deployment",
	)
	_ = deployCmd.Flags().MarkHidden("addons")

	return deployCmd
}

func NewDeployHandler() *DeployHandler {
	return &DeployHandler{}
}

type DeployHandler struct {
}

// Return the app id if it exists, otherwise return an empty string.
func (h *DeployHandler) GetAppId(ctx *CLIContext, name string) (string, error) {
	page := int64(0)
	offset := int64(0)
	limit := int64(100)

	// Consume paginated results until the application is found or the end of the list is reached.
	for {
		res, resp, err := ctx.Client.AppsApi.ListApps(ctx.Context).
			Name(name).
			Offset(strconv.FormatInt(offset, 10)).
			Limit(strconv.FormatInt(limit, 10)).
			Execute()

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
