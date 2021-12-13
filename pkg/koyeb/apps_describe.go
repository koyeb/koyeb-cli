package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper2"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Describe(cmd *cobra.Command, args []string) error {
	res, _, err := h.client.AppsApi.GetApp(h.ctxWithAuth, h.ResolveAppArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}
	resListServices, _, err := h.client.ServicesApi.ListServices(h.ctxWithAuth).AppId(res.App.GetId()).Limit("100").Execute()
	if err != nil {
		fatalApiError(err)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	describeAppsReply := NewDescribeAppReply(h.mapper, &res, full)
	listServicesReply := NewListServicesReply(&resListServices, full)

	return renderer.NewDescribeRenderer(describeAppsReply, listServicesReply).Render(output)
}

type DescribeAppReply struct {
	mapper *idmapper2.Mapper
	res    *koyeb.GetAppReply
	full   bool
}

func NewDescribeAppReply(mapper *idmapper2.Mapper, res *koyeb.GetAppReply, full bool) *DescribeAppReply {
	return &DescribeAppReply{
		mapper: mapper,
		res:    res,
		full:   full,
	}
}

func (a *DescribeAppReply) MarshalBinary() ([]byte, error) {
	return a.res.GetApp().MarshalJSON()
}

func (a *DescribeAppReply) Title() string {
	return "App"
}

func (a *DescribeAppReply) Headers() []string {
	return []string{"id", "name", "domains", "created_at", "updated_at"}
}

func (a *DescribeAppReply) Fields() []map[string]string {
	res := []map[string]string{}
	item := a.res.GetApp()
	fields := map[string]string{
		"id":         renderer.FormatID2(a.mapper, item.GetId(), a.full),
		"name":       item.GetName(),
		"domains":    formatDomains(item.GetDomains()),
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
		"updated_at": renderer.FormatTime(item.GetUpdatedAt()),
	}
	res = append(res, fields)
	return res
}
