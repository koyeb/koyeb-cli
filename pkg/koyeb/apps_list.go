package koyeb

import (
	"fmt"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper2"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *AppHandler) List(cmd *cobra.Command, args []string) error {
	results := &koyeb.ListAppsReply{}

	page := 0
	offset := 0
	limit := 100
	for {
		res, _, err := h.client.AppsApi.ListApps(h.ctxWithAuth).Limit(fmt.Sprintf("%d", limit)).Offset(fmt.Sprintf("%d", offset)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		if results.Apps == nil {
			results.Apps = res.Apps
		} else {
			*results.Apps = append(*results.Apps, *res.Apps...)
		}

		page++
		offset = page * limit
		if int64(offset) >= res.GetCount() {
			break
		}
	}

	full, _ := cmd.Flags().GetBool("full")
	listAppsReply := NewListAppsReply(h.mapper, results, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.NewListRenderer(listAppsReply).Render(output)
}

type ListAppsReply struct {
	mapper *idmapper2.Mapper
	res    *koyeb.ListAppsReply
	full   bool
}

func NewListAppsReply(mapper *idmapper2.Mapper, res *koyeb.ListAppsReply, full bool) *ListAppsReply {
	return &ListAppsReply{
		mapper: mapper,
		res:    res,
		full:   full,
	}
}

func (a *ListAppsReply) MarshalBinary() ([]byte, error) {
	return a.res.MarshalJSON()
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

	for _, item := range a.res.GetApps() {
		fields := map[string]string{
			"id": renderer.FormatID2(a.mapper, item.GetId(), a.full),
			//"id":         renderer.FormatID(item.GetId(), a.full),
			"name":       item.GetName(),
			"domains":    formatDomains(item.GetDomains()),
			"created_at": renderer.FormatTime(item.GetCreatedAt()),
		}
		res = append(res, fields)
	}
	return res
}
