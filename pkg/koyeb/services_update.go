package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Update(ctx *CLIContext, cmd *cobra.Command, args []string, updateService *koyeb.UpdateService) error {
	res, resp, err := ctx.client.ServicesApi.UpdateService(ctx.context, h.ResolveServiceArgs(ctx, args[0])).Service(*updateService).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}
	log.Infof("Service deployment in progress. Access deployment logs running: koyeb service logs %s.", res.Service.GetId()[:8])

	full := GetBoolFlags(cmd, "full")
	getServiceReply := NewGetServiceReply(ctx.mapper, &koyeb.GetServiceReply{Service: res.Service}, full)
	return ctx.renderer.Render(getServiceReply)
}
