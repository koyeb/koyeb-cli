package koyeb

import (
	"fmt"

	stderrors "errors"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Try to create a koyeb application. Do nothing if the application already exists.
func TryCreateKoyebApplication(name string, ctx *CLIContext) error {
	createApp := koyeb.NewCreateAppWithDefaults()
	createApp.SetName(DatabaseAppName)

	_, resp, err := ctx.Client.AppsApi.CreateApp(ctx.Context).App(*createApp).Execute()
	if err != nil {
		var openAPIError *koyeb.GenericOpenAPIError

		// The only way to know if the CreateApp API call failed because the app already exists is to check the error message.
		// This is not ideal, but the API does not return a specific error code for this case.
		if stderrors.As(err, &openAPIError) {
			if errorWithFields, ok := openAPIError.Model().(koyeb.ErrorWithFields); ok {
				fields := errorWithFields.GetFields()
				if len(fields) == 1 && fields[0].GetField() == "name" && fields[0].GetDescription() == "already exists" {
					return nil
				}
			}
		}
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while creating the app `%s`", name),
			err,
			resp,
		)
	}
	return nil
}

func (h *DatabaseHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string, createService *koyeb.CreateService) error {
	if err := TryCreateKoyebApplication(DatabaseAppName, ctx); err != nil {
		return err
	}

	appID, err := h.ResolveAppArgs(ctx, DatabaseAppName)
	if err != nil {
		return err
	}

	createService.SetAppId(appID)
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
