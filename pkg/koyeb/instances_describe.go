package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *InstanceHandler) Describe(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	instance, err := h.ResolveInstanceArgs(ctx, args[0])
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.InstancesApi.GetInstance(ctx.Context, instance).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the instance `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	describeInstancesReply := NewDescribeInstanceReply(ctx.Mapper, res, full)
	ctx.Renderer.Render(describeInstancesReply)
	return nil
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
		"id":         renderer.FormatID(item.GetId(), r.full),
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
