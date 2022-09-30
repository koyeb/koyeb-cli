package koyeb

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Resume(cmd *cobra.Command, args []string) error {
	_, resp, err := h.client.AppsApi.ResumeApp(h.ctx, h.ResolveAppArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	log.Infof("App %s resuming.", args[0])
	return nil
}
