package koyeb

import (
	"fmt"
	"time"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *DeploymentHandler) Logs(ctx *CLIContext, cmd *cobra.Command, since time.Time, args []string) error {
	deployment, err := h.ResolveDeploymentArgs(ctx, args[0])
	if err != nil {
		return err
	}

	deploymentDetail, resp, err := ctx.Client.DeploymentsApi.GetDeployment(ctx.Context, deployment).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the deployment `%s`", args[0]),
			err,
			resp,
		)
	}

	logsQuery, err := ctx.LogsClient.NewWatchLogsQuery(
		GetStringFlags(cmd, "type"),
		"",
		deploymentDetail.Deployment.GetId(),
		"",
		since,
		GetBoolFlags(cmd, "full"),
	)
	if err != nil {
		return err
	}
	return logsQuery.PrintAll()
}
