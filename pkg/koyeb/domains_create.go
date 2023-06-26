package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

func (h *DomainHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	createDomainReq := koyeb.NewCreateDomainWithDefaults()
	createDomainReq.SetName(args[0])
	createDomainReq.SetType(koyeb.DOMAINTYPE_CUSTOM)

	attachToApp := GetStringFlags(cmd, "attach-to")
	if attachToApp != "" {
		appID, err := ctx.Mapper.App().ResolveID(attachToApp)
		if err != nil {
			fatalApiError(err, nil)
		}

		createDomainReq.SetAppId(appID)
	}

	res, resp, err := ctx.Client.DomainsApi.CreateDomain(ctx.Context).Domain(*createDomainReq).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	getDomainsReply := NewGetDomainReply(ctx.Mapper, &koyeb.GetDomainReply{Domain: res.Domain}, full)
	return ctx.Renderer.Render(getDomainsReply)
}
