package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Update(ctx *CLIContext, cmd *cobra.Command, args []string, updateApp *koyeb.UpdateApp) error {
	app, err := h.ResolveAppArgs(ctx, args[0])
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.AppsApi.UpdateApp2(ctx.Context, app).App(*updateApp).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while updating the application `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getAppsReply := NewGetAppReply(ctx.Mapper, &koyeb.GetAppReply{App: res.App}, full)
	ctx.Renderer.Render(getAppsReply)
	return nil
}
