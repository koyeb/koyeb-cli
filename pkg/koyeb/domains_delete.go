package koyeb

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *DomainHandler) Delete(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	_, resp, err := ctx.client.DomainsApi.DeleteDomain(ctx.context, h.ResolveDomainArgs(ctx, args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}
	log.Infof("Domain %s deleted.", args[0])
	return nil
}
