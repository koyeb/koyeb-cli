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

	logsType := GetStringFlags(cmd, "type")
	startStr := GetStringFlags(cmd, "start-time")
	endStr := GetStringFlags(cmd, "end-time")
	regex := GetStringFlags(cmd, "regex-search")
	text := GetStringFlags(cmd, "text-search")
	order := GetStringFlags(cmd, "order")
	tail := GetBoolFlags(cmd, "tail")
	output := GetStringFlags(cmd, "output")

	return ctx.LogsClient.PrintLogs(ctx, LogsQuery{
		Type:         logsType,
		DeploymentId: deploymentDetail.Deployment.GetId(),
		ServiceId:    "",
		InstanceId:   "",
		Since:        since,
		Start:        startStr,
		End:          endStr,
		Text:         text,
		Order:        order,
		Tail:         tail,
		Regex:        regex,
		Full:         GetBoolFlags(cmd, "full"),
		Output:       output,
	})
}
