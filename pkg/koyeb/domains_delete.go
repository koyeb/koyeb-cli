package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *DomainHandler) Delete(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	domain, err := h.ResolveDomainArgs(ctx, args[0])
	if err != nil {
		return err
	}

	_, resp, err := ctx.Client.DomainsApi.DeleteDomain(ctx.Context, domain).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while deleting the domain `%s`", args[0]),
			err,
			resp,
		)
	}
	log.Infof("Domain %s deleted.", args[0])
	return nil
}
