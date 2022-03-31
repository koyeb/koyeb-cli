package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *InstanceHandler) Describe(cmd *cobra.Command, args []string) error {
	res, resp, err := h.client.InstancesApi.GetInstance(h.ctx, h.ResolveInstanceArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	describeInstancesReply := NewDescribeInstanceReply(h.mapper, &res, full)

	return renderer.NewDescribeRenderer(describeInstancesReply).Render(output)
}

type DescribeInstanceReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetInstanceReply
	full   bool
}

func NewDescribeInstanceReply(mapper *idmapper.Mapper, value *koyeb.GetInstanceReply, full bool) *DescribeInstanceReply {
	return &DescribeInstanceReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (DescribeInstanceReply) Title() string {
	return "Instance"
}

func (r *DescribeInstanceReply) MarshalBinary() ([]byte, error) {
	return r.value.GetInstance().MarshalJSON()
}

func (r *DescribeInstanceReply) Headers() []string {
	return []string{"id", "service", "status", "region", "datacenter", "messages", "created_at", "updated_at"}
}

func (r *DescribeInstanceReply) Fields() []map[string]string {
	item := r.value.GetInstance()
	fields := map[string]string{
		"id":         renderer.FormatInstanceID(r.mapper, item.GetId(), r.full),
		"service":    renderer.FormatServiceSlug(r.mapper, item.GetServiceId(), r.full),
		"status":     formatInstanceStatus(item.GetStatus()),
		"region":     item.GetRegion(),
		"datacenter": item.GetDatacenter(),
		"messages":   formatMessages(item.GetMessages()),
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
		"updated_at": renderer.FormatTime(item.GetUpdatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}
