package koyeb

import (
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *DeploymentHandler) List(cmd *cobra.Command, args []string) error {
	list := []koyeb.DeploymentListItem{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		res, _, err := h.client.DeploymentsApi.ListDeployments(h.ctx).
			Limit(strconv.FormatInt(limit, 10)).Offset(strconv.FormatInt(offset, 10)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		list = append(list, res.GetDeployments()...)

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	listDeploymentsReply := NewListDeploymentsReply(h.mapper, &koyeb.ListDeploymentsReply{Deployments: &list}, full)

	return renderer.NewListRenderer(listDeploymentsReply).Render(output)
}

type ListDeploymentsReply struct {
	mapper *idmapper.Mapper
	res    *koyeb.ListDeploymentsReply
	full   bool
}

func NewListDeploymentsReply(mapper *idmapper.Mapper, res *koyeb.ListDeploymentsReply, full bool) *ListDeploymentsReply {
	return &ListDeploymentsReply{
		mapper: mapper,
		res:    res,
		full:   full,
	}
}

func (a *ListDeploymentsReply) Title() string {
	return "Deployments"
}

func (a *ListDeploymentsReply) MarshalBinary() ([]byte, error) {
	return a.res.MarshalJSON()
}

func (a *ListDeploymentsReply) Headers() []string {
	return []string{"id", "service", "status", "status_message", "regions", "created_at"}
}

func (a *ListDeploymentsReply) Fields() []map[string]string {
	res := []map[string]string{}

	for _, item := range a.res.GetDeployments() {
		fields := map[string]string{
			"id":             renderer.FormatDeploymentID(a.mapper, item.GetId(), a.full),
			"service":        renderer.FormatServiceSlug(a.mapper, item.GetServiceId(), a.full),
			"status":         formatDeploymentStatus(item.State.GetStatus()),
			"status_message": item.State.GetStatusMessage(),
			"regions":        renderRegions(item.Definition.Regions),
			"created_at":     renderer.FormatTime(item.GetCreatedAt()),
		}
		res = append(res, fields)
	}

	return res
}

func renderRegions(regions *[]string) string {
	if regions == nil {
		return "-"
	}

	return strings.Join(*regions, ",")
}
