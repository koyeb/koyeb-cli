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

func (h *ServiceHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string, createService *koyeb.CreateService, wait bool) error {
	appID, err := h.parseAppName(cmd, args[0])
	if err != nil {
		return err
	}

	app, err := h.ResolveAppArgs(ctx, appID)
	if err != nil {
		return err
	}

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
		ctxd, cancel := context.WithTimeout(ctx.Context, 5*time.Minute)
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

			if res.Service != nil && res.Service.Status != nil &&
				*res.Service.Status != koyeb.SERVICESTATUS_STARTING {
				return nil
			}
		}

		log.Infof("Service deployment still in progress, --wait timed out. To access the build logs, run: `koyeb service logs %s -t build`. For the runtime logs, run `koyeb service logs %s`",
			res.Service.GetId()[:8],
			res.Service.GetId()[:8],
		)
		return nil
	}

	return nil
}
