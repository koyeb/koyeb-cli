package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

// Get retrieves a service pool by name, short ID, or full UUID.
func (h *PoolHandler) Get(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	poolID, err := ResolvePoolArgs(ctx, args[0])
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.ServicePoolsApi.GetServicePool(ctx.Context, poolID).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the pool `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getPoolReply := NewPoolReply(ctx.Mapper, res.GetServicePool(), full)
	ctx.Renderer.Render(getPoolReply)
	return nil
}
