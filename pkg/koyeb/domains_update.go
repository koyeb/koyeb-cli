package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *DomainHandler) SetSubdomain(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	updateDomainReq := koyeb.NewUpdateDomainWithDefaults()
	appID, err := ctx.Mapper.App().ResolveID(args[0])
	if err != nil {
		return err
	}
	updateDomainReq.SetAppId(appID)
	updateDomainReq.SetSubdomain(args[1])

	domainID, err := ctx.Mapper.App().GetAutoDomain(appID)
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.DomainsApi.UpdateDomain(ctx.Context, domainID).Domain(*updateDomainReq).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while renaming the automatic domain for `%s` to %q", args[0], args[1]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getDomainsReply := NewGetDomainReply(ctx.Mapper, &koyeb.GetDomainReply{Domain: res.Domain}, full)
	ctx.Renderer.Render(getDomainsReply)
	return nil
}
