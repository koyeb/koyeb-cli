package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string, createService *koyeb.CreateService) error {
	app, _ := cmd.Flags().GetString("app")
	resApp, resp, err := ctx.client.AppsApi.GetApp(ctx.context, h.ResolveAppArgs(ctx, app)).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	createService.SetAppId(resApp.App.GetId())
	res, resp, err := ctx.client.ServicesApi.CreateService(ctx.context).Service(*createService).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	log.Infof("Service deployment in progress. Access deployment logs running: koyeb service logs %s.", res.Service.GetId()[:8])

	full := GetBoolFlags(cmd, "full")
	getServiceReply := NewGetServiceReply(ctx.mapper, &koyeb.GetServiceReply{Service: res.Service}, full)
	return ctx.renderer.Render(getServiceReply)
}
