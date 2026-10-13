package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Delete deletes a sandbox by name, short ID, or full UUID. The sandbox
// type is verified before the deletion call so a regular service can never
// be destroyed by a sandbox command.
func (h *SandboxHandler) Delete(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	service, err := resolveSandboxService(ctx, args[0])
	if err != nil {
		return err
	}

	resp, err := ctx.API.DeleteService(ctx.Context, service.GetId())
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while deleting the sandbox `%s`", args[0]),
			err,
			resp,
		)
	}
	log.Infof("Sandbox %s deleted.", args[0])
	return nil
}
