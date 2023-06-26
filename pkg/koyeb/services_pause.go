package koyeb

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Pause(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	_, resp, err := ctx.Client.ServicesApi.PauseService(ctx.Context, h.ResolveServiceArgs(ctx, args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}
	log.Infof("Service %s pausing.", args[0])
	return nil
}
