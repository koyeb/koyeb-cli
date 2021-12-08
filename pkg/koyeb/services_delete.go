package koyeb

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Delete(cmd *cobra.Command, args []string) error {
	_, _, err := h.client.ServicesApi.DeleteService(h.ctxWithAuth, h.ResolveServiceShortID(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}
	log.Infof("Service %s deleted.", args[0])
	return nil
}
