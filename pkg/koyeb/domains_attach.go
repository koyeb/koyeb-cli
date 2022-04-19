package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *DomainHandler) Attach(cmd *cobra.Command, args []string) error {
	domainID, err := h.mapper.Domain().ResolveID(args[0])
	if err != nil {
		fatalApiError(err, nil)
	}

	appID, err := h.mapper.App().ResolveID(args[1])
	if err != nil {
		fatalApiError(err, nil)
	}

	updateDomainReq := koyeb.NewUpdateDomainWithDefaults()
	updateDomainReq.SetAppId(appID)

	_, resp, err := h.client.DomainsApi.UpdateDomain(h.ctx, domainID).Body(*updateDomainReq).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	res, resp, err := h.client.AppsApi.GetApp(h.ctx, appID).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")

	getAppReply := NewGetAppReply(h.mapper, &res, full)
	return renderer.NewDescribeItemRenderer(getAppReply).Render(output)
}
