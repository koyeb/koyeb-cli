package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper2"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Get(cmd *cobra.Command, args []string) error {
	res, _, err := h.client.AppsApi.GetApp(h.ctxWithAuth, h.ResolveAppArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getAppsReply := NewGetAppReply(h.mapper, &res, full)

	return renderer.NewItemRenderer(getAppsReply).Render(output)
}

type GetAppReply struct {
	mapper *idmapper2.Mapper
	res    *koyeb.GetAppReply
	full   bool
}

func NewGetAppReply(mapper *idmapper2.Mapper, res *koyeb.GetAppReply, full bool) *GetAppReply {
	return &GetAppReply{
		mapper: mapper,
		res:    res,
		full:   full,
	}
}

func (a *GetAppReply) MarshalBinary() ([]byte, error) {
	return a.res.GetApp().MarshalJSON()
}

func (a *GetAppReply) Title() string {
	return "App"
}

func (a *GetAppReply) Headers() []string {
	return []string{"id", "name", "domains", "created_at"}
}

func (a *GetAppReply) Fields() []map[string]string {
	res := []map[string]string{}
	item := a.res.GetApp()
	fields := map[string]string{
		"id":         renderer.FormatAppID(a.mapper, item.GetId(), a.full),
		"name":       item.GetName(),
		"domains":    formatDomains(item.GetDomains()),
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
	}
	res = append(res, fields)
	return res
}
