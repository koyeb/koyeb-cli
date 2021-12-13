package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Update(cmd *cobra.Command, args []string, updateApp *koyeb.UpdateApp) error {
	res, _, err := h.client.AppsApi.UpdateApp2(h.ctx, h.ResolveAppArgs(args[0])).Body(*updateApp).Execute()
	if err != nil {
		fatalApiError(err)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getAppsReply := NewGetAppReply(h.mapper, &koyeb.GetAppReply{App: res.App}, full)

	return renderer.NewDescribeItemRenderer(getAppsReply).Render(output)
}
