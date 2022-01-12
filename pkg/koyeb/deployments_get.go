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
	res, resp, err := h.client.DeploymentsApi.GetDeployment(h.ctx, h.ResolveDeploymentArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
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
	return []string{"id", "service", "status", "messages", "regions", "created_at"}
}

func (r *GetDeploymentReply) Fields() []map[string]string {
	item := r.value.GetDeployment()
	fields := map[string]string{
		"id":         renderer.FormatDeploymentID(r.mapper, item.GetId(), r.full),
		"service":    renderer.FormatServiceSlug(r.mapper, item.GetServiceId(), r.full),
		"status":     formatDeploymentStatus(item.GetStatus()),
		"messages":   formatDeploymentMessages(item.GetMessages(), 0),
		"regions":    renderRegions(item.Definition.Regions),
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}

func formatDeploymentStatus(ds koyeb.DeploymentStatus) string {
	return string(ds)
}

func formatDeploymentMessages(messages []string, max int) string {
	concat := strings.Join(messages, " ")
	if max == 0 || len(concat) < max {
		return concat
	}
	return fmt.Sprint(concat[:max], "...")
}

func renderRegions(regions *[]string) string {
	if regions == nil {
		return "-"
	}

	return strings.Join(*regions, ",")
}
