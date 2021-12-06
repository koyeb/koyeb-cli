package koyeb

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Init(cmd *cobra.Command, args []string, createApp *koyeb.CreateApp, createService *koyeb.CreateService) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	_, _, err := client.ServicesApi.CreateService(ctx).DryRun(true).Body(*createService).Execute()
	if err != nil {
		fatalApiError(err)
	}

	createApp.SetName(args[0])
	res, _, err := client.AppsApi.CreateApp(ctx).Body(*createApp).Execute()
	if err != nil {
		fatalApiError(err)
	}
	createService.SetAppId(res.App.GetId())

	serviceRes, _, err := client.ServicesApi.CreateService(ctx).Body(*createService).Execute()
	if err != nil {
		fatalApiError(err)
	}

	full, _ := cmd.Flags().GetBool("full")
	getAppsReply := NewGetAppReply(&koyeb.GetAppReply{App: res.App}, full)
	getServiceReply := NewGetServiceReply(&koyeb.GetServiceReply{Service: serviceRes.Service}, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.NewDescribeRenderer(
		getAppsReply,
		getServiceReply,
	).Render(output)
}
