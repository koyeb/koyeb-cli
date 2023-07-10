package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Describe(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	service, err := h.ResolveServiceArgs(ctx, args[0])
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.ServicesApi.GetService(ctx.Context, service).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the service `%s`", args[0]),
			err,
			resp,
		)
	}

	instancesRes, resp, err := ctx.Client.InstancesApi.ListInstances(ctx.Context).
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
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while listing the instances of the service `%s`", args[0]),
			err,
			resp,
		)
	}

	deploymentsRes, resp, err := ctx.Client.DeploymentsApi.ListDeployments(ctx.Context).ServiceId(res.Service.GetId()).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while listing the deployments of the service `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")

	getServiceReply := NewGetServiceReply(ctx.Mapper, res, full)
	listInstancesReply := NewListInstancesReply(ctx.Mapper, instancesRes, full)
	listDeploymentsReply := NewListDeploymentsReply(ctx.Mapper, deploymentsRes, full)
	renderer.NewChainRenderer(ctx.Renderer).
		Render(getServiceReply).
		Render(listInstancesReply).
		Render(listDeploymentsReply)
	return nil
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
		"id":         renderer.FormatID(item.GetId(), r.full),
		"app":        renderer.FormatAppName(r.mapper, item.GetAppId(), r.full),
		"name":       item.GetName(),
		"status":     formatServiceStatus(item.GetStatus()),
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
		"updated_at": renderer.FormatTime(item.GetUpdatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}
