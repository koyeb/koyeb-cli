package koyeb

import (
	"github.com/spf13/cobra"
)

func (h *DomainHandler) Refresh(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	id := h.ResolveDomainArgs(ctx, args[0])
	_, resp, err := ctx.Client.DomainsApi.RefreshDomainStatus(ctx.Context, id).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	res, resp, err := ctx.Client.DomainsApi.GetDomain(ctx.Context, id).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	getDomainsReply := NewGetDomainReply(ctx.Mapper, res, full)
	ctx.Renderer.Render(getDomainsReply)
	return nil
}
