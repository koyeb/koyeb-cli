package koyeb

import (
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper2"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *AppHandler) List(cmd *cobra.Command, args []string) error {
	list := []koyeb.AppListItem{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		res, _, err := h.client.AppsApi.ListApps(h.ctxWithAuth).
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
	listAppsReply := NewListAppsReply(h.mapper, list, full)

	return renderer.NewListRenderer(listAppsReply).Render(output)
}

type ListAppsReply struct {
	mapper *idmapper2.Mapper
	list   []koyeb.AppListItem
	full   bool
}

func NewListAppsReply(mapper *idmapper2.Mapper, list []koyeb.AppListItem, full bool) *ListAppsReply {
	return &ListAppsReply{
		mapper: mapper,
		list:   list,
		full:   full,
	}
}

func (a *ListAppsReply) MarshalBinary() ([]byte, error) {
	rep := &koyeb.ListAppsReply{
		Apps: &a.list,
	}
	return rep.MarshalJSON()
}

func (a *ListAppsReply) Title() string {
	return "Apps"
}

func (a *ListAppsReply) Headers() []string {
	return []string{"id", "name", "domains", "created_at"}
}

func formatDomains(domains []koyeb.Domain) string {
	strL := []string{}
	for _, d := range domains {
		strL = append(strL, d.GetName())
	}
	return strings.Join(strL, ",")
}

func (a *ListAppsReply) Fields() []map[string]string {
	res := []map[string]string{}

	for _, item := range a.list {
		fields := map[string]string{
			"id":         renderer.FormatID2(a.mapper, item.GetId(), a.full),
			"name":       item.GetName(),
			"domains":    formatDomains(item.GetDomains()),
			"created_at": renderer.FormatTime(item.GetCreatedAt()),
		}
		res = append(res, fields)
	}
	return res
}
