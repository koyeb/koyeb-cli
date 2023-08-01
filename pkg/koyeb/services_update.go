package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Update(ctx *CLIContext, cmd *cobra.Command, args []string, updateService *koyeb.UpdateService) error {
	service, err := h.ResolveServiceArgs(ctx, args[0])
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.ServicesApi.UpdateService(ctx.Context, service).Service(*updateService).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while updating the service `%s`", args[0]),
			err,
			resp,
		)
	}
	log.Infof(
		"Service deployment in progress. To access the build logs, run: `koyeb service logs %s -t build`. For the runtime logs, run `koyeb service logs %s`",
		res.Service.GetId()[:8],
		res.Service.GetId()[:8],
	)

	full := GetBoolFlags(cmd, "full")
	getServiceReply := NewGetServiceReply(ctx.Mapper, &koyeb.GetServiceReply{Service: res.Service}, full)
	ctx.Renderer.Render(getServiceReply)
	return nil
}
