package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Delete(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	app, err := h.ResolveAppArgs(ctx, args[0])
	if err != nil {
		return err
	}

	_, resp, err := ctx.Client.AppsApi.DeleteApp(ctx.Context, app).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while deleting the application `%s`", args[0]),
			err,
			resp,
		)
	}

	log.Infof("App %s deleted.", args[0])
	return nil
}
