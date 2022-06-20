package koyeb

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Resume(cmd *cobra.Command, args []string) error {
	_, resp, err := h.client.ServicesApi.ResumeService(h.ctx, h.ResolveServiceArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}
	log.Infof("Service %s resuming.", args[0])
	return nil
}
