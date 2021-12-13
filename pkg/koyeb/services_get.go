package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Get(cmd *cobra.Command, args []string) error {
	res, _, err := h.client.ServicesApi.GetService(h.ctxWithAuth, h.ResolveServiceArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getServiceReply := NewGetServiceReply(h.mapper, &res, full)

	return renderer.NewItemRenderer(getServiceReply).Render(output)
}

type GetServiceReply struct {
	mapper *idmapper.Mapper
	res    *koyeb.GetServiceReply
	full   bool
}

func NewGetServiceReply(mapper *idmapper.Mapper, res *koyeb.GetServiceReply, full bool) *GetServiceReply {
	return &GetServiceReply{
		mapper: mapper,
		res:    res,
		full:   full,
	}
}

func (a *GetServiceReply) MarshalBinary() ([]byte, error) {
	return a.res.GetService().MarshalJSON()
}

func (a *GetServiceReply) Title() string {
	return "Service"
}

func (a *GetServiceReply) Headers() []string {
	return []string{"id", "app", "name", "version", "status", "created_at"}
}

func (a *GetServiceReply) Fields() []map[string]string {
	res := []map[string]string{}
	item := a.res.GetService()
	fields := map[string]string{
		"id":         renderer.FormatServiceID(a.mapper, item.GetId(), a.full),
		"app":        renderer.FormatAppName(a.mapper, item.GetAppId(), a.full),
		"name":       item.GetName(),
		"version":    item.GetVersion(),
		"status":     formatStatus(item.State.GetStatus()),
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
	}
	res = append(res, fields)
	return res
}

func formatStatus(status koyeb.ServiceStateStatus) string {
	return fmt.Sprintf("%s", status)
}
