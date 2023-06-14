package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Describe(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	res, resp, err := ctx.client.ServicesApi.GetService(ctx.context, h.ResolveServiceArgs(ctx, args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	instancesRes, resp, err := ctx.client.InstancesApi.ListInstances(ctx.context).
		Statuses([]string{
			string(koyeb.INSTANCESTATUS_ALLOCATING),
			string(koyeb.INSTANCESTATUS_STARTING),
			string(koyeb.INSTANCESTATUS_HEALTHY),
			string(koyeb.INSTANCESTATUS_UNHEALTHY),
			string(koyeb.INSTANCESTATUS_STOPPING),
		}).
		ServiceId(res.Service.GetId()).
		Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	deploymentsRes, resp, err := ctx.client.DeploymentsApi.ListDeployments(ctx.context).ServiceId(res.Service.GetId()).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")

	getServiceReply := NewGetServiceReply(ctx.mapper, res, full)
	listInstancesReply := NewListInstancesReply(ctx.mapper, instancesRes, full)
	listDeploymentsReply := NewListDeploymentsReply(ctx.mapper, deploymentsRes, full)

	return renderer.NewDescribeRenderer(getServiceReply, listDeploymentsReply, listInstancesReply).Render(output)
}

type DescribeServiceReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetServiceReply
	full   bool
}

func NewDescribeServiceReply(mapper *idmapper.Mapper, value *koyeb.GetServiceReply, full bool) *DescribeServiceReply {
	return &DescribeServiceReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (DescribeServiceReply) Title() string {
	return "Service"
}

func (r *DescribeServiceReply) MarshalBinary() ([]byte, error) {
	return r.value.GetService().MarshalJSON()
}

func (r *DescribeServiceReply) Headers() []string {
	return []string{"id", "app", "name", "status", "created_at", "updated_at"}
}

func (r *DescribeServiceReply) Fields() []map[string]string {
	item := r.value.GetService()
	fields := map[string]string{
		"id":         renderer.FormatServiceID(r.mapper, item.GetId(), r.full),
		"app":        renderer.FormatAppName(r.mapper, item.GetAppId(), r.full),
		"name":       item.GetName(),
		"status":     formatServiceStatus(item.GetStatus()),
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
		"updated_at": renderer.FormatTime(item.GetUpdatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}
