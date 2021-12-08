package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *DeploymentHandler) Describe(cmd *cobra.Command, args []string) error {
	ctx := h.ctxWithAuth
	res, _, err := h.client.DeploymentsApi.GetDeployment(ctx, h.ResolveDeploymentShortID(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	instancesRes, _, err := h.client.InstancesApi.ListInstances(ctx).Statuses([]string{
		string(koyeb.INSTANCESTATUS_ALLOCATING),
		string(koyeb.INSTANCESTATUS_STARTING),
		string(koyeb.INSTANCESTATUS_HEALTHY),
		string(koyeb.INSTANCESTATUS_UNHEALTHY),
		string(koyeb.INSTANCESTATUS_STOPPING),
	}).DeploymentId(res.Deployment.GetId()).Execute()
	if err != nil {
		fatalApiError(err)
	}

	appMapper := idmapper.NewAppMapper(ctx, h.client)
	serviceMapper := idmapper.NewServiceMapper(ctx, h.client)

	full, _ := cmd.Flags().GetBool("full")
	describeDeploymentsReply := NewDescribeDeploymentReply(&res, full)
	defDeployment := renderer.NewGenericRenderer("Definition", res.Deployment.Definition)
	listInstancesReply := NewListInstancesReply(instancesRes, appMapper, serviceMapper, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.NewMultiRenderer(
		renderer.NewDescribeRenderer(describeDeploymentsReply),
		renderer.NewSeparatorRenderer(),
		defDeployment,
		renderer.NewSeparatorRenderer(),
		renderer.NewTitleRenderer(listInstancesReply),
		renderer.NewListRenderer(listInstancesReply),
	).Render(output)
}

type DescribeDeploymentReply struct {
	res  *koyeb.GetDeploymentReply
	full bool
}

func NewDescribeDeploymentReply(res *koyeb.GetDeploymentReply, full bool) *DescribeDeploymentReply {
	return &DescribeDeploymentReply{
		res:  res,
		full: full,
	}
}

func (a *DescribeDeploymentReply) MarshalBinary() ([]byte, error) {
	return a.res.GetDeployment().MarshalJSON()
}

func (a *DescribeDeploymentReply) Title() string {
	return "Deployment"
}

func (a *DescribeDeploymentReply) Headers() []string {
	return []string{"id", "app", "service", "status", "status_message", "created_at", "updated_at"}
}

func (a *DescribeDeploymentReply) Fields() []map[string]string {
	res := []map[string]string{}
	item := a.res.GetDeployment()
	fields := map[string]string{
		"id":             renderer.FormatID(item.GetId(), a.full),
		"app":            renderer.FormatID(item.GetAppId(), a.full),
		"service":        renderer.FormatID(item.GetServiceId(), a.full),
		"status":         formatDeploymentStatus(item.State.GetStatus()),
		"status_message": item.State.GetStatusMessage(),
		"created_at":     renderer.FormatTime(item.GetCreatedAt()),
		"updated_at":     renderer.FormatTime(item.GetUpdatedAt()),
	}
	res = append(res, fields)
	return res
}
