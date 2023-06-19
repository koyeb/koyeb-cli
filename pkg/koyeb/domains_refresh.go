package koyeb

import (
	"github.com/spf13/cobra"
)

func (h *DomainHandler) Refresh(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	id := h.ResolveDomainArgs(ctx, args[0])
	_, resp, err := ctx.client.DomainsApi.RefreshDomainStatus(ctx.context, id).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	res, resp, err := ctx.client.DomainsApi.GetDomain(ctx.context, id).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	getDomainsReply := NewGetDomainReply(ctx.mapper, res, full)
	return ctx.renderer.Render(getDomainsReply)
}
