package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Create(cmd *cobra.Command, args []string, createService *koyeb.CreateService) error {
	app, _ := cmd.Flags().GetString("app")
	resApp, resp, err := h.client.AppsApi.GetApp(h.ctx, h.ResolveAppArgs(app)).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	createService.SetAppId(resApp.App.GetId())
	res, resp, err := h.client.ServicesApi.CreateService(h.ctx).Body(*createService).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	log.Infof("Service deployment in progress. Access deployment logs running: koyeb service logs %s.", res.Service.GetId()[:8])

	wait := GetBoolFlags(cmd, "wait")
	if wait {
		watchDeployment(h, res.Service.GetLatestDeploymentId())
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")

	gRes, gResp, err := h.client.ServicesApi.GetService(h.ctx, res.Service.GetId()).Execute()
	if err != nil {
		fatalApiError(err, gResp)
	}

	getServiceReply := NewGetServiceReply(h.mapper, &koyeb.GetServiceReply{Service: gRes.Service}, full)

	return renderer.NewDescribeRenderer(getServiceReply).Render(output)
}
