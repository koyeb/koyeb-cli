package koyeb

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Init(ctx *CLIContext, cmd *cobra.Command, args []string, createApp *koyeb.CreateApp, createService *koyeb.CreateService) error {
	wait, _ := cmd.Flags().GetBool("wait")
	waitTimeout, err := cmd.Flags().GetDuration("wait-timeout")
	if err != nil {
		return err
	}

	uid := uuid.Must(uuid.NewV4())
	createService.SetAppId(uid.String())
	_, resp, err := ctx.Client.ServicesApi.CreateService(ctx.Context).DryRun(true).Service(*createService).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while creating the service `%s`", args[0]),
			err,
			resp,
		)
	}

	createApp.SetName(args[0])
	res, resp, err := ctx.Client.AppsApi.CreateApp(ctx.Context).App(*createApp).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while creating the application `%s`", args[0]),
			err,
			resp,
		)
	}
	createService.SetAppId(res.App.GetId())

	serviceRes, resp, err := ctx.Client.ServicesApi.CreateService(ctx.Context).Service(*createService).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the service `%s`", args[0]),
			err,
			resp,
		)
	}
	defer func() {
		res, _, err := ctx.Client.AppsApi.GetApp(ctx.Context, res.App.GetId()).Execute()
		if err != nil {
			return
		}
		app := res.App
		getServiceRes, _, err := ctx.Client.ServicesApi.GetService(ctx.Context, serviceRes.Service.GetId()).Execute()
		if err != nil {
			return
		}
		full := GetBoolFlags(cmd, "full")
		getAppsReply := NewGetAppReply(ctx.Mapper, &koyeb.GetAppReply{App: app}, full)
		getServiceReply := NewGetServiceReply(ctx.Mapper, &koyeb.GetServiceReply{Service: getServiceRes.Service}, full)
		renderer.NewChainRenderer(ctx.Renderer).Render(getAppsReply).Render(getServiceReply)
	}()

	if wait {
		log.Infof("App deployment is in progress. To access the build logs, run: `koyeb service logs %s -t build`. For the runtime logs, run `koyeb service logs %s`",
			serviceRes.Service.GetId()[:8],
			serviceRes.Service.GetId()[:8],
		)

		ctxd, cancel := context.WithTimeout(ctx.Context, waitTimeout)
		defer cancel()

		for range ticker(ctxd, 2*time.Second) {
			getServiceRes, resp, err := ctx.Client.ServicesApi.GetService(ctxd, serviceRes.Service.GetId()).Execute()
			if err != nil {
				return errors.NewCLIErrorFromAPIError(
					"Error while fetching service",
					err,
					resp,
				)
			}

			if getServiceRes.Service != nil && getServiceRes.Service.Status != nil {
				switch status := *getServiceRes.Service.Status; status {
				case koyeb.SERVICESTATUS_DELETED, koyeb.SERVICESTATUS_DEGRADED, koyeb.SERVICESTATUS_UNHEALTHY:
					return fmt.Errorf("service %s deployment ended in status: %s", serviceRes.Service.GetId()[:8], status)
				case koyeb.SERVICESTATUS_STARTING, koyeb.SERVICESTATUS_RESUMING, koyeb.SERVICESTATUS_DELETING, koyeb.SERVICESTATUS_PAUSING:
					break
				default:
					return nil
				}
			}
		}

		log.Infof("Service deployment still in progress, --wait timed out. To access the build logs, run: `koyeb service logs %s -t build`. For the runtime logs, run `koyeb service logs %s`",
			serviceRes.Service.GetId()[:8],
			serviceRes.Service.GetId()[:8],
		)
		return fmt.Errorf("service deployment still in progress, --wait timed out. To access the build logs, run: `koyeb service logs %s -t build`. For the runtime logs, run `koyeb service logs %s`",
			serviceRes.Service.GetId()[:8],
			serviceRes.Service.GetId()[:8],
		)
	}

	return nil
}
