package koyeb

import (
	"fmt"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *InstanceHandler) Get(cmd *cobra.Command, args []string) error {
	res, resp, err := h.client.InstancesApi.GetInstance(h.ctx, h.ResolveInstanceArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getInstancesReply := NewGetInstanceReply(h.mapper, &res, full)

	return renderer.NewItemRenderer(getInstancesReply).Render(output)
}

type GetInstanceReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetInstanceReply
	full   bool
}

func NewGetInstanceReply(mapper *idmapper.Mapper, value *koyeb.GetInstanceReply, full bool) *GetInstanceReply {
	return &GetInstanceReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (GetInstanceReply) Title() string {
	return "Instance"
}

func (r *GetInstanceReply) MarshalBinary() ([]byte, error) {
	return r.value.GetInstance().MarshalJSON()
}

func (r *GetInstanceReply) Headers() []string {
	return []string{"id", "service", "status", "deployment_id", "datacenter", "created_at"}
}

func (r *GetInstanceReply) Fields() []map[string]string {
	item := r.value.GetInstance()
	fields := map[string]string{
		"id":            renderer.FormatInstanceID(r.mapper, item.GetId(), r.full),
		"service":       renderer.FormatServiceSlug(r.mapper, item.GetServiceId(), r.full),
		"status":        formatInstanceStatus(item.GetStatus()),
		"deployment_id": renderer.FormatDeploymentID(r.mapper, item.GetDeploymentId(), r.full),
		"datacenter":    item.GetDatacenter(),
		"created_at":    renderer.FormatTime(item.GetCreatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}

func formatInstanceStatus(status koyeb.InstanceStatus) string {
	return fmt.Sprintf("%s", status)
}

func formatMessages(msg []string) string {
	return strings.Join(msg, "\n")
}
