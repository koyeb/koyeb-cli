package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Get(cmd *cobra.Command, args []string) error {
	res, resp, err := h.client.ServicesApi.GetService(h.ctx, h.ResolveServiceArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getServiceReply := NewGetServiceReply(h.mapper, res, full)

	return renderer.NewItemRenderer(getServiceReply).Render(output)
}

type GetServiceReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetServiceReply
	full   bool
}

func NewGetServiceReply(mapper *idmapper.Mapper, value *koyeb.GetServiceReply, full bool) *GetServiceReply {
	return &GetServiceReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (GetServiceReply) Title() string {
	return "Service"
}

func (r *GetServiceReply) MarshalBinary() ([]byte, error) {
	return r.value.GetService().MarshalJSON()
}

func (r *GetServiceReply) Headers() []string {
	return []string{"id", "app", "name", "status", "created_at"}
}

func (r *GetServiceReply) Fields() []map[string]string {
	item := r.value.GetService()
	fields := map[string]string{
		"id":         renderer.FormatServiceID(r.mapper, item.GetId(), r.full),
		"app":        renderer.FormatAppName(r.mapper, item.GetAppId(), r.full),
		"name":       item.GetName(),
		"status":     formatServiceStatus(item.GetStatus()),
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}

func formatServiceStatus(status koyeb.ServiceStatus) string {
	return string(status)
}
