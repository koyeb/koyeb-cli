package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Delete deletes a service pool by name, short ID, or full UUID.
func (h *PoolHandler) Delete(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	pool, err := ResolvePoolArgs(ctx, args[0])
	if err != nil {
		return err
	}

	resp, err := ctx.API.DeleteServicePool(ctx.Context, pool)
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while deleting the pool `%s`", args[0]),
			err,
			resp,
		)
	}
	log.Infof("Pool %s deleted.", args[0])
	return nil
}
