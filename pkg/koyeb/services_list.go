package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) List(cmd *cobra.Command, args []string) error {
	results := koyeb.ListServicesReply{}

	page := 0
	offset := 0
	limit := 100
	for {
		req := h.client.ServicesApi.ListServices(h.ctxWithAuth)
		appId, _ := cmd.Flags().GetString("app")
		if appId != "" {
			req = req.AppId(h.ResolveAppShortID(appId))
		}
		name, _ := cmd.Flags().GetString("name")
		if name != "" {
			req = req.Name(name)
		}
		res, _, err := req.Limit(fmt.Sprintf("%d", limit)).Offset(fmt.Sprintf("%d", offset)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		if results.Services == nil {
			results = res
		} else {
			*results.Services = append(*results.Services, *res.Services...)
		}

		page += 1
		offset = page * limit
		if int64(offset) >= res.GetCount() {
			break
		}
	}

	full, _ := cmd.Flags().GetBool("full")
	listServicesReply := NewListServicesReply(&results, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.NewListRenderer(listServicesReply).Render(output)
}

type ListServicesReply struct {
	res  *koyeb.ListServicesReply
	full bool
}

func NewListServicesReply(res *koyeb.ListServicesReply, full bool) *ListServicesReply {
	return &ListServicesReply{
		res:  res,
		full: full,
	}
}

func (a *ListServicesReply) Title() string {
	return "Services"
}

func (a *ListServicesReply) MarshalBinary() ([]byte, error) {
	return a.res.MarshalJSON()
}

func (a *ListServicesReply) Headers() []string {
	return []string{"id", "app", "name", "status", "created_at"}
}

func (a *ListServicesReply) Fields() []map[string]string {
	res := []map[string]string{}

	for _, item := range a.res.GetServices() {
		fields := map[string]string{
			"id":         renderer.FormatID(item.GetId(), a.full),
			"app":        renderer.FormatID(item.GetAppId(), a.full),
			"name":       item.GetName(),
			"status":     formatStatus(item.State.GetStatus()),
			"created_at": renderer.FormatTime(item.GetCreatedAt()),
		}
		res = append(res, fields)
	}
	return res
}
