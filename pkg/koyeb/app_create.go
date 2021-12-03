package koyeb

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Create(cmd *cobra.Command, args []string, createApp *koyeb.CreateApp) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	createApp.SetName(args[0])
	res, _, err := client.AppsApi.CreateApp(ctx).Body(*createApp).Execute()
	if err != nil {
		fatalApiError(err)
	}
	full, _ := cmd.Flags().GetBool("full")
	getAppsReply := NewGetAppReply(&koyeb.GetAppReply{App: res.App}, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.DescribeRenderer(output, getAppsReply)
}
