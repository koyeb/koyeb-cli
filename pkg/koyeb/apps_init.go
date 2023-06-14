package koyeb

import (
	"github.com/gofrs/uuid"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Init(ctx *CLIContext, cmd *cobra.Command, args []string, createApp *koyeb.CreateApp, createService *koyeb.CreateService) error {
	uid := uuid.Must(uuid.NewV4())
	createService.SetAppId(uid.String())
	_, resp, err := ctx.client.ServicesApi.CreateService(ctx.context).DryRun(true).Service(*createService).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	createApp.SetName(args[0])
	res, resp, err := ctx.client.AppsApi.CreateApp(ctx.context).App(*createApp).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}
	createService.SetAppId(res.App.GetId())

	serviceRes, resp, err := ctx.client.ServicesApi.CreateService(ctx.context).Service(*createService).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getAppsReply := NewGetAppReply(ctx.mapper, &koyeb.GetAppReply{App: res.App}, full)
	getServiceReply := NewGetServiceReply(ctx.mapper, &koyeb.GetServiceReply{Service: serviceRes.Service}, full)

	return renderer.NewDescribeRenderer(getAppsReply, getServiceReply).Render(output)
}
