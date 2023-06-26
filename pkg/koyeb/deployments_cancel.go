package koyeb

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *DeploymentHandler) Cancel(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	_, resp, err := ctx.Client.DeploymentsApi.CancelDeployment(ctx.Context, h.ResolveDeploymentArgs(ctx, args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}
	log.Infof("Deployment %s canceled.", args[0])
	return nil
}
