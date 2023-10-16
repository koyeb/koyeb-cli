package koyeb

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

func (h *InstanceHandler) Exec(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	instance, err := h.ResolveInstanceArgs(ctx, args[0])
	if err != nil {
		return err
	}

	returnCode, err := ctx.ExecClient.Exec(ctx.Context, ExecId{
		Id:   instance,
		Type: koyeb.EXECCOMMANDREQUESTIDTYPE_INSTANCE_ID,
	}, args[1:])
	if err != nil {
		return &errors.CLIError{
			What:       "Error while executing the command",
			Why:        "the CLI did not succeed to execute the command",
			Additional: nil,
			Orig:       err,
			Solution:   "Make sure the command is correct and exists in the service. If the problem persists, try to update the CLI to the latest version.",
		}
	}
	if returnCode != 0 {
		os.Exit(returnCode)
	}
	return nil
}
