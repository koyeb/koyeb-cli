package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *AppHandler) CreateApp(ctx *CLIContext, payload *koyeb.CreateApp) (*koyeb.CreateAppReply, error) {
	res, resp, err := ctx.Client.AppsApi.CreateApp(ctx.Context).App(*payload).Execute()
	if err != nil {
		return nil, errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while creating the app `%s`", payload.GetName()),
			err,
			resp,
		)
	}
	return res, nil
}

func (h *AppHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string, createApp *koyeb.CreateApp) error {
	createApp.SetName(args[0])

	res, err := h.CreateApp(ctx, createApp)
	if err != nil {
		return err
	}

	full := GetBoolFlags(cmd, "full")
	getAppsReply := NewGetAppReply(ctx.Mapper, &koyeb.GetAppReply{App: res.App}, full)
	ctx.Renderer.Render(getAppsReply)
	return nil
}
