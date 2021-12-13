package koyeb

import (
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *AppHandler) List(cmd *cobra.Command, args []string) error {
	list := []koyeb.AppListItem{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		res, _, err := h.client.AppsApi.ListApps(h.ctx).
			Limit(strconv.FormatInt(limit, 10)).Offset(strconv.FormatInt(offset, 10)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		list = append(list, res.GetApps()...)

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	listAppsReply := NewListAppsReply(h.mapper, &koyeb.ListAppsReply{Apps: &list}, full)

	return renderer.NewListRenderer(listAppsReply).Render(output)
}

type ListAppsReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.ListAppsReply
	full   bool
}

func NewListAppsReply(mapper *idmapper.Mapper, value *koyeb.ListAppsReply, full bool) *ListAppsReply {
	return &ListAppsReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (ListAppsReply) Title() string {
	return "Apps"
}

func (r *ListAppsReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *ListAppsReply) Headers() []string {
	return []string{"id", "name", "domains", "created_at"}
}

func (r *ListAppsReply) Fields() []map[string]string {
	items := r.value.GetApps()
	resp := make([]map[string]string, 0, len(items))

	for _, item := range items {
		fields := map[string]string{
			"id":         renderer.FormatAppID(r.mapper, item.GetId(), r.full),
			"name":       item.GetName(),
			"domains":    formatDomains(item.GetDomains()),
			"created_at": renderer.FormatTime(item.GetCreatedAt()),
		}
		resp = append(resp, fields)
	}

	return resp
}
