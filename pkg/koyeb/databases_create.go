package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *DatabaseHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string, createService *koyeb.CreateService) error {
	appID, err := parseAppName(cmd, args[0])
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
			"Error while creating the database service",
			err,
			resp,
		)
	}

	log.Infof(
		"Database creation in progress. To access the connection strings, run `koyeb database get %s` in a few seconds.",
		res.Service.GetId()[:8],
	)
	return nil
}
