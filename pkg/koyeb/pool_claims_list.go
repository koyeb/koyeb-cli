package koyeb

import (
	"fmt"
	"strconv"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func newPoolClaimsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list POOL",
		Short: "List claims of a service pool",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			if err := setProjectHeader(ctx, cmd); err != nil {
				return err
			}
			poolID, err := ResolvePoolArgs(ctx, args[0])
			if err != nil {
				return err
			}

			req := ctx.Client.PoolClaimsApi.ListClaim(ctx.Context, poolID)

			if cmd.Flags().Changed("status") {
				status, _ := cmd.Flags().GetString("status")
				req = req.Status(status)
			}
			if cmd.Flags().Changed("limit") {
				limit, _ := cmd.Flags().GetInt64("limit")
				req = req.Limit(strconv.FormatInt(limit, 10))
			}
			if cmd.Flags().Changed("offset") {
				offset, _ := cmd.Flags().GetInt64("offset")
				req = req.Offset(strconv.FormatInt(offset, 10))
			}

			res, resp, err := req.Execute()
			if err != nil {
				return errors.NewCLIErrorFromAPIError(
					fmt.Sprintf("Error while listing the claims of the pool `%s`", args[0]),
					err,
					resp,
				)
			}

			full := GetBoolFlags(cmd, "full")
			listPoolClaimsReply := NewListPoolClaimsReply(ctx.Mapper, res, full)
			ctx.Renderer.Render(listPoolClaimsReply)
			return nil
		}),
	}
	cmd.Flags().String("status", "", "Filter claims by status")
	cmd.Flags().Int64("limit", 0, "Limit the number of claims returned")
	cmd.Flags().Int64("offset", 0, "Offset the claims returned")

	return cmd
}
