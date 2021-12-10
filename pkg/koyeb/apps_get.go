package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Get(cmd *cobra.Command, args []string) error {
	res, _, err := h.client.AppsApi.GetApp(h.ctxWithAuth, h.ResolveAppArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	full, _ := cmd.Flags().GetBool("full")
	getAppsReply := NewGetAppReply(&res, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.NewItemRenderer(getAppsReply).Render(output)
}

type GetAppReply struct {
	res  *koyeb.GetAppReply
	full bool
}

func NewGetAppReply(res *koyeb.GetAppReply, full bool) *GetAppReply {
	return &GetAppReply{
		res:  res,
		full: full,
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
		"id":         renderer.FormatID(item.GetId(), a.full),
		"name":       item.GetName(),
		"domains":    formatDomains(item.GetDomains()),
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
	}
	res = append(res, fields)
	return res
}
