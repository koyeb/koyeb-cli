package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Create(cmd *cobra.Command, args []string, createApp *koyeb.CreateApp) error {
	createApp.SetName(args[0])
	res, _, err := h.client.AppsApi.CreateApp(h.ctxWithAuth).Body(*createApp).Execute()
	if err != nil {
		fatalApiError(err)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getAppsReply := NewGetAppReply(h.mapper, &koyeb.GetAppReply{App: res.App}, full)

	return renderer.NewDescribeItemRenderer(getAppsReply).Render(output)
}
