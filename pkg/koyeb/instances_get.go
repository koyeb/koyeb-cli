package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *InstanceHandler) Get(cmd *cobra.Command, args []string) error {
	res, _, err := h.client.InstancesApi.GetInstance(h.ctxWithAuth, h.ResolveInstanceShortID(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	full, _ := cmd.Flags().GetBool("full")
	getInstancesReply := NewGetInstanceReply(&res, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.NewItemRenderer(getInstancesReply).Render(output)
}

type GetInstanceReply struct {
	res  *koyeb.GetInstanceReply
	full bool
}

func NewGetInstanceReply(res *koyeb.GetInstanceReply, full bool) *GetInstanceReply {
	return &GetInstanceReply{
		res:  res,
		full: full,
	}
}

func (a *GetInstanceReply) MarshalBinary() ([]byte, error) {
	return a.res.GetInstance().MarshalJSON()
}

func (a *GetInstanceReply) Title() string {
	return "Instance"
}

func (a *GetInstanceReply) Headers() []string {
	return []string{"id", "service", "status", "deployment_id", "datacenter", "created_at"}
}

func (a *GetInstanceReply) Fields() []map[string]string {
	res := []map[string]string{}
	item := a.res.GetInstance()
	fields := map[string]string{
		"id":            renderer.FormatID(item.GetId(), a.full),
		"service":       renderer.FormatID(item.GetServiceId(), a.full),
		"status":        formatInstanceStatus(item.GetStatus()),
		"deployment_id": renderer.FormatID(item.GetDeploymentId(), a.full),
		"datacenter":    item.GetDatacenter(),
		"created_at":    renderer.FormatTime(item.GetCreatedAt()),
	}
	res = append(res, fields)
	return res
}
