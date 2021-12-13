package koyeb

import (
	"fmt"
	"strings"

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
	value  *koyeb.GetDeploymentReply
	full   bool
}

func NewGetDeploymentReply(mapper *idmapper.Mapper, value *koyeb.GetDeploymentReply, full bool) *GetDeploymentReply {
	return &GetDeploymentReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (GetDeploymentReply) Title() string {
	return "Deployment"
}

func (r *GetDeploymentReply) MarshalBinary() ([]byte, error) {
	return r.value.GetDeployment().MarshalJSON()
}

func (r *GetDeploymentReply) Headers() []string {
	return []string{"id", "service", "status", "status_message", "regions", "created_at"}
}

func (r *GetDeploymentReply) Fields() []map[string]string {
	item := r.value.GetDeployment()
	fields := map[string]string{
		"id":             renderer.FormatDeploymentID(r.mapper, item.GetId(), r.full),
		"service":        renderer.FormatServiceSlug(r.mapper, item.GetServiceId(), r.full),
		"status":         formatDeploymentStatus(item.State.GetStatus()),
		"status_message": item.State.GetStatusMessage(),
		"regions":        renderRegions(item.Definition.Regions),
		"created_at":     renderer.FormatTime(item.GetCreatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}

func formatDeploymentStatus(ds koyeb.ServiceRevisionStateStatus) string {
	return fmt.Sprintf("%s", ds)
}

func renderRegions(regions *[]string) string {
	if regions == nil {
		return "-"
	}

	return strings.Join(*regions, ",")
}
