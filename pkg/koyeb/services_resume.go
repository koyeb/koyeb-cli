package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Resume(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	serviceName, err := parseServiceName(cmd, args[0])
	if err != nil {
		return err
	}

	service, err := h.ResolveServiceArgs(ctx, serviceName)
	if err != nil {
		return err
	}

	_, resp, err := ctx.Client.ServicesApi.ResumeService(ctx.Context, service).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while resuming the service `%s`", serviceName),
			err,
			resp,
		)
	}
	log.Infof("Service %s resuming.", serviceName)
	return nil
}
