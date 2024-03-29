package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *DomainHandler) Detach(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	domainID, err := ctx.Mapper.Domain().ResolveID(args[0])
	if err != nil {
		return err
	}

	updateDomainReq := koyeb.NewUpdateDomainWithDefaults()
	updateDomainReq.SetAppId("")
	res, resp, err := ctx.Client.DomainsApi.UpdateDomain(ctx.Context, domainID).Domain(*updateDomainReq).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while updating the domain `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")

	getDomainReply := NewGetDomainReply(ctx.Mapper, &koyeb.GetDomainReply{Domain: res.Domain}, full)
	ctx.Renderer.Render(getDomainReply)
	return nil
}
