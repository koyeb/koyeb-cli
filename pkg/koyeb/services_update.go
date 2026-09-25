package koyeb

import (
	"fmt"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Update(ctx *CLIContext, cmd *cobra.Command, args []string, updateService *koyeb.UpdateService) error {

	serviceName, err := h.parseServiceName(cmd, args[0])
	if err != nil {
		return err
	}

	service, err := h.ResolveServiceArgs(ctx, serviceName)
	if err != nil {
		return err
	}

	wait, _ := cmd.Flags().GetBool("wait")

	res, resp, err := ctx.Client.ServicesApi.UpdateService(ctx.Context, service).Service(*updateService).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while updating the service `%s`", serviceName),
			err,
			resp,
		)
	}
	log.Infof(
		"Service deployment in progress. To access the build logs, run: `koyeb service logs %s -t build`. For the runtime logs, run `koyeb service logs %s`",
		res.Service.GetId()[:8],
		res.Service.GetId()[:8],
	)

	defer func() {
		res, _, err := ctx.Client.ServicesApi.GetService(ctx.Context, res.Service.GetId()).Execute()
		if err != nil {
			return
		}
		full := GetBoolFlags(cmd, "full")
		getServiceReply := NewGetServiceReply(ctx.Mapper, &koyeb.GetServiceReply{Service: res.Service}, full)
		ctx.Renderer.Render(getServiceReply)
	}()

	if wait {
		waitTimeout, err := waitTimeoutFlag(cmd)
		if err != nil {
			return err
		}
		if err := waitEngine(ctx.Context, waitTimeout, 2*time.Second,
			deploymentUpdateProbe(ctx, res.Service.GetLatestDeploymentId()),
			func() error { return serviceWaitTimedOut(res.Service.GetId(), "service logs") },
		); err != nil {
			return err
		}
	}
	return nil
}
