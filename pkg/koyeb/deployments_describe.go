package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *DeploymentHandler) Describe(cmd *cobra.Command, args []string) error {
	res, _, err := h.client.DeploymentsApi.GetDeployment(h.ctx, h.ResolveDeploymentArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	instancesRes, _, err := h.client.InstancesApi.ListInstances(h.ctx).
		Statuses([]string{
			string(koyeb.INSTANCESTATUS_ALLOCATING),
			string(koyeb.INSTANCESTATUS_STARTING),
			string(koyeb.INSTANCESTATUS_HEALTHY),
			string(koyeb.INSTANCESTATUS_UNHEALTHY),
			string(koyeb.INSTANCESTATUS_STOPPING),
		}).
		DeploymentId(res.Deployment.GetId()).
		Execute()
	if err != nil {
		fatalApiError(err)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")

	describeDeploymentsReply := NewDescribeDeploymentReply(h.mapper, &res, full)
	defDeployment := renderer.NewGenericRenderer("Definition", res.Deployment.Definition)
	listInstancesReply := NewListInstancesReply(h.mapper, &instancesRes, full)

	return renderer.
		NewMultiRenderer(
			renderer.NewDescribeRenderer(describeDeploymentsReply),
			renderer.NewSeparatorRenderer(),
			defDeployment,
			renderer.NewSeparatorRenderer(),
			renderer.NewTitleRenderer(listInstancesReply),
			renderer.NewListRenderer(listInstancesReply),
		).
		Render(output)
}

type DescribeDeploymentReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetDeploymentReply
	full   bool
}

func NewDescribeDeploymentReply(mapper *idmapper.Mapper, value *koyeb.GetDeploymentReply, full bool) *DescribeDeploymentReply {
	return &DescribeDeploymentReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (DescribeDeploymentReply) Title() string {
	return "Deployment"
}

func (r *DescribeDeploymentReply) MarshalBinary() ([]byte, error) {
	return r.value.GetDeployment().MarshalJSON()
}

func (r *DescribeDeploymentReply) Headers() []string {
	return []string{"id", "service", "status", "status_message", "regions", "created_at", "updated_at"}
}

func (r *DescribeDeploymentReply) Fields() []map[string]string {
	item := r.value.GetDeployment()
	fields := map[string]string{
		"id":             renderer.FormatDeploymentID(r.mapper, item.GetId(), r.full),
		"service":        renderer.FormatServiceSlug(r.mapper, item.GetServiceId(), r.full),
		"status":         formatDeploymentStatus(item.State.GetStatus()),
		"status_message": item.State.GetStatusMessage(),
		"regions":        renderRegions(item.Definition.Regions),
		"created_at":     renderer.FormatTime(item.GetCreatedAt()),
		"updated_at":     renderer.FormatTime(item.GetUpdatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}
