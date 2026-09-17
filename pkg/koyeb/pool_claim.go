package koyeb

import (
	"fmt"
	"uuid"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

// generateRequestID returns a UUID v7 used when --request-id is not given.
func generateRequestID() string {
	return uuid.NewV7().String()
}

// buildPoolClaimRequest builds the PoolClaimRequest sent to the claim API.
func buildPoolClaimRequest(poolID, requestID string) koyeb.PoolClaimRequest {
	return koyeb.PoolClaimRequest{
		PoolId:    &poolID,
		RequestId: &requestID,
	}
}

func newPoolClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim POOL",
		Short: "Claim an instance from a pool",
		Long: `Claim an instance from a pool.

When --request-id is not provided, a random UUID v7 is generated. Replaying
a claim with the same request ID is idempotent: the server returns the
previously created claim instead of provisioning a new instance.`,
		Args: cobra.ExactArgs(1),
		Example: `
# Claim an instance from a pool
$> koyeb pool claim my-pool

# Claim with an explicit request ID (idempotent retries)
$> koyeb pool claim my-pool --request-id my-request-id
`,
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			if err := setProjectHeader(ctx, cmd); err != nil {
				return err
			}
			poolID, err := ResolvePoolArgs(ctx, args[0])
			if err != nil {
				return err
			}

			requestID, _ := cmd.Flags().GetString("request-id")
			if requestID == "" {
				requestID = generateRequestID()
			}

			req := buildPoolClaimRequest(poolID, requestID)

			res, resp, err := ctx.Client.PoolClaimsApi.Claim(ctx.Context).Body(req).Execute()
			if err != nil {
				return errors.NewCLIErrorFromAPIError(
					fmt.Sprintf("Error while claiming an instance from the pool `%s`", args[0]),
					err,
					resp,
				)
			}

			full := GetBoolFlags(cmd, "full")
			claimReply := NewClaimReply(ctx.Mapper, res, full)
			ctx.Renderer.Render(claimReply)
			return nil
		}),
	}
	cmd.Flags().String("request-id", "", "Claim request ID (defaults to a generated UUID v4)")

	return cmd
}
