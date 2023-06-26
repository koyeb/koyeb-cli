package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

func (h *DomainHandler) Attach(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	domainID, err := ctx.Mapper.Domain().ResolveID(args[0])
	if err != nil {
		fatalApiError(err, nil)
	}

	appID, err := ctx.Mapper.App().ResolveID(args[1])
	if err != nil {
		fatalApiError(err, nil)
	}

	updateDomainReq := koyeb.NewUpdateDomainWithDefaults()
	updateDomainReq.SetAppId(appID)

	_, resp, err := ctx.Client.DomainsApi.UpdateDomain(ctx.Context, domainID).Domain(*updateDomainReq).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	res, resp, err := ctx.Client.AppsApi.GetApp(ctx.Context, appID).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")

	getAppReply := NewGetAppReply(ctx.Mapper, res, full)
	return ctx.Renderer.Render(getAppReply)
}
