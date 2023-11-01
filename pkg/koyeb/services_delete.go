package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Delete(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	serviceName, err := parseServiceName(cmd, args[0])
	if err != nil {
		return err
	}

	service, err := h.ResolveServiceArgs(ctx, serviceName)
	if err != nil {
		return err
	}

	_, resp, err := ctx.Client.ServicesApi.DeleteService(ctx.Context, service).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while deleting the service `%s`", serviceName),
			err,
			resp,
		)
	}
	log.Infof("Service %s deleted.", serviceName)
	return nil
}
