package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *DomainHandler) Refresh(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	domain, err := h.ResolveDomainArgs(ctx, args[0])
	if err != nil {
		return err
	}

	_, resp, err := ctx.Client.DomainsApi.RefreshDomainStatus(ctx.Context, domain).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while refreshing the status of the domain `%s`", domain),
			err,
			resp,
		)
	}

	res, resp, err := ctx.Client.DomainsApi.GetDomain(ctx.Context, domain).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the domain `%s`", domain),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getDomainsReply := NewGetDomainReply(ctx.Mapper, res, full)
	ctx.Renderer.Render(getDomainsReply)
	return nil
}
