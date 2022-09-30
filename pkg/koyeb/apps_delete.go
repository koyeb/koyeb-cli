package koyeb

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Delete(cmd *cobra.Command, args []string) error {
	_, resp, err := h.client.AppsApi.DeleteApp(h.ctx, h.ResolveAppArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	log.Infof("App %s deleted.", args[0])
	return nil
}
