package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *DeploymentHandler) Get(cmd *cobra.Command, args []string) error {
	res, _, err := h.client.DeploymentsApi.GetDeployment(h.ctxWithAuth, h.ResolveDeploymentShortID(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	full, _ := cmd.Flags().GetBool("full")
	getDeploymentsReply := NewGetDeploymentReply(&res, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.NewItemRenderer(getDeploymentsReply).Render(output)
}

type GetDeploymentReply struct {
	res  *koyeb.GetDeploymentReply
	full bool
}

func NewGetDeploymentReply(res *koyeb.GetDeploymentReply, full bool) *GetDeploymentReply {
	return &GetDeploymentReply{
		res:  res,
		full: full,
	}
}

func (a *GetDeploymentReply) MarshalBinary() ([]byte, error) {
	return a.res.GetDeployment().MarshalJSON()
}

func (a *GetDeploymentReply) Title() string {
	return "Deployment"
}

func (a *GetDeploymentReply) Headers() []string {
	return []string{"id", "service", "status", "status_message", "regions", "created_at"}
}

func (a *GetDeploymentReply) Fields() []map[string]string {
	res := []map[string]string{}
	item := a.res.GetDeployment()
	fields := map[string]string{
		"id":             renderer.FormatID(item.GetId(), a.full),
		"service":        renderer.FormatID(item.GetServiceId(), a.full),
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
