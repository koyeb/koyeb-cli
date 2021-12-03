package koyeb

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Update(cmd *cobra.Command, args []string, updateApp *koyeb.UpdateApp) error {
	client := getApiClient()
	ctx := getAuth(context.Background())
	res, _, err := client.AppsApi.UpdateApp2(ctx, args[0]).Body(*updateApp).Execute()
	if err != nil {
		fatalApiError(err)
	}
	full, _ := cmd.Flags().GetBool("full")
	getAppsReply := NewGetAppReply(&koyeb.GetAppReply{App: res.App}, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.DescribeRenderer(output, getAppsReply)
}