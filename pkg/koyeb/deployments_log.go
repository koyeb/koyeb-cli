package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

func (h *DeploymentHandler) Log(cmd *cobra.Command, args []string) error {
	deploymentDetail, _, err := h.client.DeploymentsApi.GetDeployment(h.ctxWithAuth, h.ResolveDeploymentShortID(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	logType, _ := cmd.Flags().GetString("type")

	done := make(chan struct{})

	query := &watchLogQuery{deploymentID: koyeb.PtrString(deploymentDetail.Deployment.GetId())}
	if logType != "" {
		query.logType = koyeb.PtrString(logType)
	}
	return watchLog(query, done)
}
