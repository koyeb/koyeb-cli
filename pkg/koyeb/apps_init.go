package koyeb

import (
	"github.com/gofrs/uuid"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Init(cmd *cobra.Command, args []string, createApp *koyeb.CreateApp, createService *koyeb.CreateService) error {
	uid := uuid.Must(uuid.NewV4())
	createService.SetAppId(uid.String())
	_, resp, err := h.client.ServicesApi.CreateService(h.ctx).DryRun(true).Body(*createService).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	createApp.SetName(args[0])
	res, resp, err := h.client.AppsApi.CreateApp(h.ctx).Body(*createApp).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}
	createService.SetAppId(res.App.GetId())

	serviceRes, resp, err := h.client.ServicesApi.CreateService(h.ctx).Body(*createService).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getAppsReply := NewGetAppReply(h.mapper, &koyeb.GetAppReply{App: res.App}, full)
	getServiceReply := NewGetServiceReply(h.mapper, &koyeb.GetServiceReply{Service: serviceRes.Service}, full)

	return renderer.NewDescribeRenderer(getAppsReply, getServiceReply).Render(output)
}
