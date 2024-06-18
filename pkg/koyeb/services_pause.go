package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Pause(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	serviceName, err := h.parseServiceName(cmd, args[0])
	if err != nil {
		return err
	}

	service, err := h.ResolveServiceArgs(ctx, serviceName)
	if err != nil {
		return err
	}

	_, resp, err := ctx.Client.ServicesApi.PauseService(ctx.Context, service).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while pausing the service `%s`", serviceName),
			err,
			resp,
		)
	}
	log.Infof("Service %s pausing.", serviceName)
	return nil
}
