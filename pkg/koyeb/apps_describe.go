package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Describe(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	res, resp, err := ctx.client.AppsApi.GetApp(ctx.context, h.ResolveAppArgs(ctx, args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}
	resListServices, resp, err := ctx.client.ServicesApi.ListServices(ctx.context).AppId(res.App.GetId()).Limit("100").Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	describeAppsReply := NewDescribeAppReply(ctx.mapper, res, full)
	listServicesReply := NewListServicesReply(ctx.mapper, resListServices, full)
	return renderer.NewChainRenderer(ctx.renderer).Render(describeAppsReply).Render(listServicesReply).Err()
}

type DescribeAppReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetAppReply
	full   bool
}

func NewDescribeAppReply(mapper *idmapper.Mapper, value *koyeb.GetAppReply, full bool) *DescribeAppReply {
	return &DescribeAppReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (DescribeAppReply) Title() string {
	return "App"
}

func (r *DescribeAppReply) MarshalBinary() ([]byte, error) {
	return r.value.GetApp().MarshalJSON()
}

func (r *DescribeAppReply) Headers() []string {
	return []string{"id", "name", "status", "domains", "created_at", "updated_at"}
}

func (r *DescribeAppReply) Fields() []map[string]string {
	item := r.value.GetApp()
	fields := map[string]string{
		"id":         renderer.FormatAppID(r.mapper, item.GetId(), r.full),
		"name":       item.GetName(),
		"status":     formatAppStatus(item.GetStatus()),
		"domains":    formatDomains(item.GetDomains(), 0),
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
		"updated_at": renderer.FormatTime(item.GetUpdatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}
