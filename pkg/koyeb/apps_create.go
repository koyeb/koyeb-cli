package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string, createApp *koyeb.CreateApp) error {
	createApp.SetName(args[0])
	res, resp, err := ctx.client.AppsApi.CreateApp(ctx.context).App(*createApp).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	getAppsReply := NewGetAppReply(ctx.mapper, &koyeb.GetAppReply{App: res.App}, full)
	return ctx.renderer.Render(getAppsReply)
}
