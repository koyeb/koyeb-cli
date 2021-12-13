package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper2"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Describe(cmd *cobra.Command, args []string) error {
	ctx := h.ctxWithAuth
	res, _, err := h.client.ServicesApi.GetService(ctx, h.ResolveServiceArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	instancesRes, _, err := h.client.InstancesApi.ListInstances(ctx).Statuses([]string{
		string(koyeb.INSTANCESTATUS_ALLOCATING),
		string(koyeb.INSTANCESTATUS_STARTING),
		string(koyeb.INSTANCESTATUS_HEALTHY),
		string(koyeb.INSTANCESTATUS_UNHEALTHY),
		string(koyeb.INSTANCESTATUS_STOPPING),
	}).ServiceId(res.Service.GetId()).Execute()
	if err != nil {
		fatalApiError(err)
	}

	deploymentsRes, _, err := h.client.DeploymentsApi.ListDeployments(ctx).ServiceId(res.Service.GetId()).Execute()
	if err != nil {
		fatalApiError(err)
	}

	appMapper := idmapper.NewAppMapper(ctx, h.client)
	serviceMapper := idmapper.NewServiceMapper(ctx, h.client)

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")

	getServiceReply := NewGetServiceReply(h.mapper, &res, full)
	listInstancesReply := NewListInstancesReply(instancesRes, appMapper, serviceMapper, full)
	listDeploymentsReply := NewListDeploymentsReply(&deploymentsRes, full)

	return renderer.NewDescribeRenderer(getServiceReply, listDeploymentsReply, listInstancesReply).Render(output)
}

type DescribeServiceReply struct {
	mapper *idmapper2.Mapper
	res    *koyeb.GetServiceReply
	full   bool
}

func NewDescribeServiceReply(mapper *idmapper2.Mapper, res *koyeb.GetServiceReply, full bool) *DescribeServiceReply {
	return &DescribeServiceReply{
		mapper: mapper,
		res:    res,
		full:   full,
	}
}

func (a *DescribeServiceReply) MarshalBinary() ([]byte, error) {
	return a.res.GetService().MarshalJSON()
}

func (a *DescribeServiceReply) Title() string {
	return "Service"
}

func (a *DescribeServiceReply) Headers() []string {
	return []string{"id", "app", "name", "version", "status", "created_at", "updated_at"}
}

func (a *DescribeServiceReply) Fields() []map[string]string {
	res := []map[string]string{}
	item := a.res.GetService()
	fields := map[string]string{
		"id":         renderer.FormatServiceID(a.mapper, item.GetId(), a.full),
		"app":        renderer.FormatAppName(a.mapper, item.GetAppId(), a.full),
		"name":       item.GetName(),
		"version":    item.GetVersion(),
		"status":     formatStatus(item.State.GetStatus()),
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
		"updated_at": renderer.FormatTime(item.GetUpdatedAt()),
	}
	res = append(res, fields)
	return res
}
