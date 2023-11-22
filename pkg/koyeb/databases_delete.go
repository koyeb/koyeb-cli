package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *DatabaseHandler) Delete(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	service, err := h.ResolveDatabaseArgs(ctx, args[0])
	if err != nil {
		return err
	}

	_, resp, err := ctx.Client.ServicesApi.DeleteService(ctx.Context, service).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while deleting the database `%s`", args[0]),
			err,
			resp,
		)
	}
	log.Infof("Database %s deleted.", service)
	return nil
}
