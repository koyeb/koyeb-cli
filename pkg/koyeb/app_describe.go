package koyeb

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Describe(cmd *cobra.Command, args []string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	res, _, err := client.AppsApi.GetApp(ctx, ResolveAppShortID(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}
	resListServices, _, err := client.ServicesApi.ListServices(ctx).AppId(res.App.GetId()).Limit("100").Execute()
	if err != nil {
		fatalApiError(err)
	}
	full, _ := cmd.Flags().GetBool("full")
	describeAppsReply := NewDescribeAppReply(&res, full)
	listServicesReply := NewListServicesReply(&resListServices, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.MultiRenderer(
		func() error { return renderer.DescribeRenderer(output, describeAppsReply) }, func() error { return renderer.ListRenderer(output, listServicesReply) })
}

type DescribeAppReply struct {
	res  *koyeb.GetAppReply
	full bool
}

func NewDescribeAppReply(res *koyeb.GetAppReply, full bool) *DescribeAppReply {
	return &DescribeAppReply{
		res:  res,
		full: full,
	}
}

func (a *DescribeAppReply) MarshalBinary() ([]byte, error) {
	return a.res.GetApp().MarshalJSON()
}

func (a *DescribeAppReply) Title() string {
	return "App"
}

func (a *DescribeAppReply) Headers() []string {
	return []string{"id", "name", "domains", "updated_at"}
}

func (a *DescribeAppReply) Fields() []map[string]string {
	res := []map[string]string{}
	item := a.res.GetApp()
	fields := map[string]string{
		"id":         renderer.FormatID(item.GetId(), a.full),
		"name":       item.GetName(),
		"domains":    formatDomains(item.GetDomains()),
		"updated_at": renderer.FormatTime(item.GetUpdatedAt()),
	}
	res = append(res, fields)
	return res
}
