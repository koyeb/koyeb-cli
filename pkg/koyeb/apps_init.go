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
	_, _, err := h.client.ServicesApi.CreateService(h.ctxWithAuth).DryRun(true).Body(*createService).Execute()
	if err != nil {
		fatalApiError(err)
	}

	createApp.SetName(args[0])
	res, _, err := h.client.AppsApi.CreateApp(h.ctxWithAuth).Body(*createApp).Execute()
	if err != nil {
		fatalApiError(err)
	}
	createService.SetAppId(res.App.GetId())

	serviceRes, _, err := h.client.ServicesApi.CreateService(h.ctxWithAuth).Body(*createService).Execute()
	if err != nil {
		fatalApiError(err)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getAppsReply := NewGetAppReply(h.mapper, &koyeb.GetAppReply{App: res.App}, full)
	getServiceReply := NewGetServiceReply(&koyeb.GetServiceReply{Service: serviceRes.Service}, full)

	return renderer.NewDescribeRenderer(getAppsReply, getServiceReply).Render(output)
}
