package koyeb

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Delete(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	_, resp, err := ctx.client.ServicesApi.DeleteService(ctx.context, h.ResolveServiceArgs(ctx, args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}
	log.Infof("Service %s deleted.", args[0])
	return nil
}
