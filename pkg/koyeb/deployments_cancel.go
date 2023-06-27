package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *DeploymentHandler) Cancel(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	deployment, err := h.ResolveDeploymentArgs(ctx, args[0])
	if err != nil {
		return err
	}

	_, resp, err := ctx.Client.DeploymentsApi.CancelDeployment(ctx.Context, deployment).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while canceling the deployment `%s`", args[0]),
			err,
			resp,
		)
	}
	log.Infof("Deployment %s canceled.", args[0])
	return nil
}
