package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Resume(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	app, err := h.ResolveAppArgs(ctx, args[0])
	if err != nil {
		return err
	}

	_, resp, err := ctx.Client.AppsApi.ResumeApp(ctx.Context, app).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while resuming the application `%s`", args[0]),
			err,
			resp,
		)
	}

	log.Infof("App %s resuming.", args[0])
	return nil
}
