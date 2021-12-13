package koyeb

import (
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *InstanceHandler) Describe(cmd *cobra.Command, args []string) error {
	res, _, err := h.client.InstancesApi.GetInstance(h.ctx, h.ResolveInstanceArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	describeInstancesReply := NewDescribeInstanceReply(h.mapper, &res, full)

	return renderer.NewMultiRenderer(renderer.NewDescribeRenderer(describeInstancesReply)).Render(output)
}

type DescribeInstanceReply struct {
	mapper *idmapper.Mapper
	res    *koyeb.GetInstanceReply
	full   bool
}

func NewDescribeInstanceReply(mapper *idmapper.Mapper, res *koyeb.GetInstanceReply, full bool) *DescribeInstanceReply {
	return &DescribeInstanceReply{
		mapper: mapper,
		res:    res,
		full:   full,
	}
}

func (DescribeInstanceReply) Title() string {
	return "Instance"
}

func (r *DescribeInstanceReply) MarshalBinary() ([]byte, error) {
	return r.res.GetInstance().MarshalJSON()
}

func (r *DescribeInstanceReply) Headers() []string {
	return []string{"id", "service", "status", "deployment_id", "datacenter", "messages", "created_at", "updated_at"}
}

func (r *DescribeInstanceReply) Fields() []map[string]string {
	item := r.res.GetInstance()
	fields := map[string]string{
		"id":            renderer.FormatInstanceID(r.mapper, item.GetId(), r.full),
		"service":       renderer.FormatServiceSlug(r.mapper, item.GetServiceId(), r.full),
		"status":        formatInstanceStatus(item.GetStatus()),
		"deployment_id": renderer.FormatDeploymentID(r.mapper, item.GetDeploymentId(), r.full),
		"datacenter":    item.GetDatacenter(),
		"messages":      formatMessages(item.GetMessages()),
		"created_at":    renderer.FormatTime(item.GetCreatedAt()),
		"updated_at":    renderer.FormatTime(item.GetUpdatedAt()),
	}

	res := []map[string]string{fields}
	return res
}

func formatMessages(msg []string) string {
	return strings.Join(msg, "\n")
}
