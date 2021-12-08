package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Get(cmd *cobra.Command, args []string) error {
	res, _, err := h.client.ServicesApi.GetService(h.ctxWithAuth, h.ResolveServiceShortID(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}
	full, _ := cmd.Flags().GetBool("full")
	getServiceReply := NewGetServiceReply(&res, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.NewItemRenderer(getServiceReply).Render(output)
}

type GetServiceReply struct {
	res  *koyeb.GetServiceReply
	full bool
}

func NewGetServiceReply(res *koyeb.GetServiceReply, full bool) *GetServiceReply {
	return &GetServiceReply{
		res:  res,
		full: full,
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
		"id":         renderer.FormatID(item.GetId(), a.full),
		"app":        renderer.FormatID(item.GetAppId(), a.full),
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
