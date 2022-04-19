package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *DomainHandler) Create(cmd *cobra.Command, args []string) error {
	createDomainReq := koyeb.NewCreateDomainWithDefaults()
	createDomainReq.SetName(args[0])
	createDomainReq.SetType(koyeb.DOMAINTYPE_CUSTOM)

	attachToApp := GetStringFlags(cmd, "attach-to")
	if attachToApp != "" {
		appID, err := h.mapper.App().ResolveID(attachToApp)
		if err != nil {
			fatalApiError(err, nil)
		}

		createDomainReq.SetAppId(appID)
	}

	res, resp, err := h.client.DomainsApi.CreateDomain(h.ctx).Body(*createDomainReq).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")

	getDomainsReply := NewGetDomainReply(h.mapper, &koyeb.GetDomainReply{Domain: res.Domain}, full)
	return renderer.NewDescribeItemRenderer(getDomainsReply).Render(output)
}
