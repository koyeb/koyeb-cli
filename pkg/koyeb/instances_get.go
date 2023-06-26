package koyeb

import (
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *InstanceHandler) Get(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	res, resp, err := ctx.Client.InstancesApi.GetInstance(ctx.Context, h.ResolveInstanceArgs(ctx, args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	getInstancesReply := NewGetInstanceReply(ctx.Mapper, res, full)
	return ctx.Renderer.Render(getInstancesReply)
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
	return []string{"id", "service", "status", "region", "datacenter", "created_at"}
}

func (r *GetInstanceReply) Fields() []map[string]string {
	item := r.value.GetInstance()
	fields := map[string]string{
		"id":         renderer.FormatInstanceID(r.mapper, item.GetId(), r.full),
		"service":    renderer.FormatServiceSlug(r.mapper, item.GetServiceId(), r.full),
		"status":     formatInstanceStatus(item.GetStatus()),
		"region":     item.GetRegion(),
		"datacenter": item.GetDatacenter(),
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}

func formatInstanceStatus(status koyeb.InstanceStatus) string {
	return string(status)
}

func formatMessages(msg []string) string {
	return strings.Join(msg, "\n")
}
