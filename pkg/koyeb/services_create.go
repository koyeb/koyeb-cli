package koyeb

import (
	"context"
	"fmt"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string, createService *koyeb.CreateService) error {
	appID, err := h.parseAppName(cmd, args[0])
	if err != nil {
		return err
	}

	app, err := h.ResolveAppArgs(ctx, appID)
	if err != nil {
		return err
	}

	wait, _ := cmd.Flags().GetBool("wait")
	waitTimeout, _ := cmd.Flags().GetDuration("wait-timeout")

	resApp, resp, err := ctx.Client.AppsApi.GetApp(ctx.Context, app).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the application `%s`", appID),
			err,
			resp,
		)
	}

	createService.SetAppId(resApp.App.GetId())
	res, resp, err := ctx.Client.ServicesApi.CreateService(ctx.Context).Service(*createService).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			"Error while creating the service",
			err,
			resp,
		)
	}
	log.Infof(
		"Service deployment in progress. To access the build logs, run: `koyeb service logs %s -t build`. For the runtime logs, run `koyeb service logs %s`",
		res.Service.GetId()[:8],
		res.Service.GetId()[:8],
	)
	defer func() {
		res, _, err := ctx.Client.ServicesApi.GetService(ctx.Context, res.Service.GetId()).Execute()
		if err != nil {
			return
		}
		full := GetBoolFlags(cmd, "full")
		getServiceReply := NewGetServiceReply(ctx.Mapper, &koyeb.GetServiceReply{Service: res.Service}, full)
		ctx.Renderer.Render(getServiceReply)
	}()

	if wait {
		ctxd, cancel := context.WithTimeout(ctx.Context, waitTimeout)
		defer cancel()

		for range ticker(ctxd, 2*time.Second) {
			res, resp, err := ctx.Client.ServicesApi.GetService(ctxd, res.Service.GetId()).Execute()
			if err != nil {
				return errors.NewCLIErrorFromAPIError(
					"Error while fetching service",
					err,
					resp,
				)
			}

			if res.Service != nil && res.Service.Status != nil {
				switch status := *res.Service.Status; status {
				case koyeb.SERVICESTATUS_DELETED, koyeb.SERVICESTATUS_DEGRADED, koyeb.SERVICESTATUS_UNHEALTHY:
					return fmt.Errorf("service %s deployment ended in status: %s", res.Service.GetId()[:8], status)
				case koyeb.SERVICESTATUS_STARTING, koyeb.SERVICESTATUS_RESUMING, koyeb.SERVICESTATUS_DELETING, koyeb.SERVICESTATUS_PAUSING:
					break
				default:
					return nil
				}
			}
		}

		log.Infof("Service deployment still in progress, --wait timed out. To access the build logs, run: `koyeb service logs %s -t build`. For the runtime logs, run `koyeb service logs %s`",
			res.Service.GetId()[:8],
			res.Service.GetId()[:8],
		)
		return fmt.Errorf("service deployment still in progress, --wait timed out. To access the build logs, run: `koyeb service logs %s -t build`. For the runtime logs, run `koyeb service logs %s`",
			res.Service.GetId()[:8],
			res.Service.GetId()[:8],
		)
	}

	return nil
}
