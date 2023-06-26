package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

func (h *DomainHandler) Detach(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	domainID, err := ctx.Mapper.Domain().ResolveID(args[0])
	if err != nil {
		fatalApiError(err, nil)
	}

	updateDomainReq := koyeb.NewUpdateDomainWithDefaults()
	updateDomainReq.SetAppId("")
	res, resp, err := ctx.Client.DomainsApi.UpdateDomain(ctx.Context, domainID).Domain(*updateDomainReq).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")

	getDomainReply := NewGetDomainReply(ctx.Mapper, &koyeb.GetDomainReply{Domain: res.Domain}, full)
	return ctx.Renderer.Render(getDomainReply)
}
