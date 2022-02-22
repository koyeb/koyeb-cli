package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) ReDeploy(cmd *cobra.Command, args []string) error {
	redeployBody := *koyeb.NewRedeployRequestInfoWithDefaults()
	_, resp, err := h.client.ServicesApi.ReDeploy(h.ctx, h.ResolveServiceArgs(args[0])).Body(redeployBody).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}
	log.Infof("Service %s redeployed.", args[0])

	gRes, gResp, err := h.client.ServicesApi.GetService(h.ctx, h.ResolveServiceArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err, gResp)
	}

	wait := GetBoolFlags(cmd, "wait")
	if wait {
		watchDeployment(h, gRes.Service.GetLatestDeploymentId())
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")

	gRes, gResp, err = h.client.ServicesApi.GetService(h.ctx, h.ResolveServiceArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err, gResp)
	}

	getServiceReply := NewGetServiceReply(h.mapper, &koyeb.GetServiceReply{Service: gRes.Service}, full)

	return renderer.NewDescribeRenderer(getServiceReply).Render(output)
}
