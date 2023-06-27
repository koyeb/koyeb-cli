package koyeb

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Init(ctx *CLIContext, cmd *cobra.Command, args []string, createApp *koyeb.CreateApp, createService *koyeb.CreateService) error {
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

	full := GetBoolFlags(cmd, "full")
	getAppsReply := NewGetAppReply(ctx.Mapper, &koyeb.GetAppReply{App: res.App}, full)
	getServiceReply := NewGetServiceReply(ctx.Mapper, &koyeb.GetServiceReply{Service: serviceRes.Service}, full)
	renderer.NewChainRenderer(ctx.Renderer).Render(getAppsReply).Render(getServiceReply)
	return nil
}
