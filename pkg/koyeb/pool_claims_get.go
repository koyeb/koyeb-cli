package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func newPoolClaimsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get CLAIM",
		Short: "Get a pool claim",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			res, resp, err := ctx.Client.PoolClaimsApi.GetClaim(ctx.Context, args[0]).Execute()
			if err != nil {
				return errors.NewCLIErrorFromAPIError(
					fmt.Sprintf("Error while retrieving the claim `%s`", args[0]),
					err,
					resp,
				)
			}

			full := GetBoolFlags(cmd, "full")
			getPoolClaimReply := NewGetPoolClaimReply(ctx.Mapper, res, full)
			ctx.Renderer.Render(getPoolClaimReply)
			return nil
		}),
	}

	return cmd
}
