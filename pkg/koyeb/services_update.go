package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Update(cmd *cobra.Command, args []string, updateService *koyeb.UpdateService) error {
	res, _, err := h.client.ServicesApi.UpdateService(h.ctxWithAuth, h.ResolveServiceArgs(args[0])).Body(*updateService).Execute()
	if err != nil {
		fatalApiError(err)
	}
	log.Infof("Service deployment in progress. Access deployment logs running: koyeb service logs %s.", res.Service.GetId()[:8])

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getServiceReply := NewGetServiceReply(h.mapper, &koyeb.GetServiceReply{Service: res.Service}, full)

	return renderer.NewDescribeItemRenderer(getServiceReply).Render(output)
}
