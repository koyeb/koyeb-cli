package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Create(cmd *cobra.Command, args []string, createService *koyeb.CreateService) error {
	app, _ := cmd.Flags().GetString("app")
	resApp, _, err := h.client.AppsApi.GetApp(h.ctxWithAuth, h.ResolveAppArgs(app)).Execute()
	if err != nil {
		fatalApiError(err)
	}

	// TODO handle both notations (<app>:<sevice> and --app=app)
	createService.SetAppId(resApp.App.GetId())
	res, _, err := h.client.ServicesApi.CreateService(h.ctxWithAuth).Body(*createService).Execute()
	if err != nil {
		fatalApiError(err)
	}

	log.Infof("Service deployment in progress. Access deployment logs running: koyeb service logs %s.", res.Service.GetId()[:8])

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getServiceReply := NewGetServiceReply(h.mapper, &koyeb.GetServiceReply{Service: res.Service}, full)

	return renderer.NewDescribeRenderer(getServiceReply).Render(output)
}
