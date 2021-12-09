package koyeb

import (
	"fmt"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *DeploymentHandler) List(cmd *cobra.Command, args []string) error {
	results := koyeb.ListDeploymentsReply{}

	page := 0
	offset := 0
	limit := 100
	for {
		res, _, err := h.client.DeploymentsApi.ListDeployments(h.ctxWithAuth).Limit(fmt.Sprintf("%d", limit)).Offset(fmt.Sprintf("%d", offset)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		if results.Deployments == nil {
			results = res
		} else {
			*results.Deployments = append(*results.Deployments, *res.Deployments...)
		}

		page += 1
		offset = page * limit
		if int64(offset) >= res.GetCount() {
			break
		}
	}
	full, _ := cmd.Flags().GetBool("full")
	listDeploymentsReply := NewListDeploymentsReply(&results, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.NewListRenderer(listDeploymentsReply).Render(output)
}

type ListDeploymentsReply struct {
	res  *koyeb.ListDeploymentsReply
	full bool
}

func NewListDeploymentsReply(res *koyeb.ListDeploymentsReply, full bool) *ListDeploymentsReply {
	return &ListDeploymentsReply{
		res:  res,
		full: full,
	}
}

func (a *ListDeploymentsReply) MarshalBinary() ([]byte, error) {
	return a.res.MarshalJSON()
}

func (a *ListDeploymentsReply) Title() string {
	return "Deployments"
}

func (a *ListDeploymentsReply) Headers() []string {
	return []string{"id", "service", "status", "status_message", "regions", "created_at"}
}

func (a *ListDeploymentsReply) Fields() []map[string]string {
	res := []map[string]string{}

	for _, item := range a.res.GetDeployments() {
		fields := map[string]string{
			"id":             renderer.FormatID(item.GetId(), a.full),
			"service":        renderer.FormatID(item.GetServiceId(), a.full),
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
