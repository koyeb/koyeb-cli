package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

func (h *DeploymentHandler) Logs(cmd *cobra.Command, args []string) error {
	deploymentDetail, resp, err := h.client.DeploymentsApi.GetDeployment(h.ctx, h.ResolveDeploymentArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	done := make(chan struct{})

	logType := GetStringFlags(cmd, "type")

	query := &WatchLogQuery{}
	query.DeploymentID = koyeb.PtrString(deploymentDetail.Deployment.GetId())
	if logType != "" {
		query.LogType = koyeb.PtrString(logType)
	}

	return WatchLog(query, done)
}
