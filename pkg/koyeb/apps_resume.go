package koyeb

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Resume(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	_, resp, err := ctx.client.AppsApi.ResumeApp(ctx.context, h.ResolveAppArgs(ctx, args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	log.Infof("App %s resuming.", args[0])
	return nil
}
