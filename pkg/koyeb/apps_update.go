package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Update(ctx *CLIContext, cmd *cobra.Command, args []string, updateApp *koyeb.UpdateApp) error {
	res, resp, err := ctx.Client.AppsApi.UpdateApp2(ctx.Context, h.ResolveAppArgs(ctx, args[0])).App(*updateApp).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	getAppsReply := NewGetAppReply(ctx.Mapper, &koyeb.GetAppReply{App: res.App}, full)
	return ctx.Renderer.Render(getAppsReply)
}
