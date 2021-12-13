package koyeb

import (
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper2"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) List(cmd *cobra.Command, args []string) error {
	list := []koyeb.ServiceListItem{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		req := h.client.ServicesApi.ListServices(h.ctxWithAuth)
		appId := GetStringFlags(cmd, "app")
		if appId != "" {
			req = req.AppId(h.ResolveAppArgs(appId))
		}
		name := GetStringFlags(cmd, "name")
		if name != "" {
			req = req.Name(name)
		}
		res, _, err := req.Limit(strconv.FormatInt(limit, 10)).Offset(strconv.FormatInt(offset, 10)).Execute()
		if err != nil {
			fatalApiError(err)
		}

		list = append(list, res.GetServices()...)

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	listServicesReply := NewListServicesReply(h.mapper, &koyeb.ListServicesReply{Services: &list}, full)

	return renderer.NewListRenderer(listServicesReply).Render(output)
}

type ListServicesReply struct {
	mapper *idmapper2.Mapper
	res    *koyeb.ListServicesReply
	full   bool
}

func NewListServicesReply(mapper *idmapper2.Mapper, res *koyeb.ListServicesReply, full bool) *ListServicesReply {
	return &ListServicesReply{
		mapper: mapper,
		res:    res,
		full:   full,
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
			"id":         renderer.FormatServiceID(a.mapper, item.GetId(), a.full),
			"app":        renderer.FormatAppName(a.mapper, item.GetAppId(), a.full),
			"name":       item.GetName(),
			"status":     formatStatus(item.State.GetStatus()),
			"created_at": renderer.FormatTime(item.GetCreatedAt()),
		}
		res = append(res, fields)
	}

	return res
}
