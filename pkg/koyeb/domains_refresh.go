package koyeb

import (
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *DomainHandler) Refresh(cmd *cobra.Command, args []string) error {
	id := h.ResolveDomainArgs(args[0])
	_, resp, err := h.client.DomainsApi.RefreshDomainStatus(h.ctx, id).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	res, resp, err := h.client.DomainsApi.GetDomain(h.ctx, id).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getDomainsReply := NewGetDomainReply(h.mapper, &res, full)

	return renderer.NewItemRenderer(getDomainsReply).Render(output)
}
