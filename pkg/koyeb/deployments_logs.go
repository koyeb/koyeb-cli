package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *DeploymentHandler) Logs(ctx *CLIContext, cmd *cobra.Command, args []string) error {
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

	done := make(chan struct{})

	logType := GetStringFlags(cmd, "type")
	if logType != "" && logType != "build" && logType != "runtime" {
		return &errors.CLIError{
			What: "Error while fetching the logs",
			Why:  "the log type you provided is invalid",
			Additional: []string{
				fmt.Sprintf("The log type should be either `build` or `runtime`, not `%s`", logType),
			},
			Orig:     nil,
			Solution: "Fix the log type and try again",
		}
	}

	query := &WatchLogQuery{}
	query.DeploymentID = koyeb.PtrString(deploymentDetail.Deployment.GetId())
	if logType != "" {
		query.LogType = koyeb.PtrString(logType)
	}

	return WatchLog(query, done)
}
