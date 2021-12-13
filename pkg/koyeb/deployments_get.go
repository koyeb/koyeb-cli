package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *DeploymentHandler) Get(cmd *cobra.Command, args []string) error {
	res, _, err := h.client.DeploymentsApi.GetDeployment(h.ctx, h.ResolveDeploymentArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getDeploymentsReply := NewGetDeploymentReply(h.mapper, &res, full)

	return renderer.NewItemRenderer(getDeploymentsReply).Render(output)
}

type GetDeploymentReply struct {
	mapper *idmapper.Mapper
	res    *koyeb.GetDeploymentReply
	full   bool
}

func NewGetDeploymentReply(mapper *idmapper.Mapper, res *koyeb.GetDeploymentReply, full bool) *GetDeploymentReply {
	return &GetDeploymentReply{
		mapper: mapper,
		res:    res,
		full:   full,
	}
}

func (a *GetDeploymentReply) Title() string {
	return "Deployment"
}

func (a *GetDeploymentReply) MarshalBinary() ([]byte, error) {
	return a.res.GetDeployment().MarshalJSON()
}

func (a *GetDeploymentReply) Headers() []string {
	return []string{"id", "service", "status", "status_message", "regions", "created_at"}
}

func (a *GetDeploymentReply) Fields() []map[string]string {
	res := []map[string]string{}
	item := a.res.GetDeployment()
	fields := map[string]string{
		"id":             renderer.FormatDeploymentID(a.mapper, item.GetId(), a.full),
		"service":        renderer.FormatServiceSlug(a.mapper, item.GetServiceId(), a.full),
		"status":         formatDeploymentStatus(item.State.GetStatus()),
		"status_message": item.State.GetStatusMessage(),
		"regions":        renderRegions(item.Definition.Regions),
		"created_at":     renderer.FormatTime(item.GetCreatedAt()),
	}
	res = append(res, fields)
	return res
}

func formatDeploymentStatus(ds koyeb.ServiceRevisionStateStatus) string {
	return fmt.Sprintf("%s", ds)
}
