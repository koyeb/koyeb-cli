package koyeb

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *DeploymentHandler) Cancel(cmd *cobra.Command, args []string) error {
	_, resp, err := h.client.DeploymentsApi.CancelDeployment(h.ctx, h.ResolveDeploymentArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}
	log.Infof("Deployment %s canceled.", args[0])
	return nil
}
