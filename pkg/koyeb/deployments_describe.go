package koyeb

import (
	"github.com/ghodss/yaml"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *DeploymentHandler) Describe(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	res, resp, err := ctx.client.DeploymentsApi.GetDeployment(ctx.context, h.ResolveDeploymentArgs(ctx, args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	// TODO(tleroux): Experimental for now.
	regionalRes, resp, err := ctx.client.RegionalDeploymentsApi.ListRegionalDeployments(ctx.context).
		DeploymentId(res.Deployment.GetId()).
		Execute()
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
		DeploymentId(res.Deployment.GetId()).
		Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")

	describeDeploymentsReply := NewDescribeDeploymentReply(ctx.mapper, res, full)
	listInstancesReply := NewListInstancesReply(ctx.mapper, instancesRes, full)
	listRegionalDeploymentsReply := NewListRegionalDeploymentsReply(ctx.mapper, regionalRes, full)
	deploymentDefinitionReply := NewDescribeDeploymentDefinitionReply(res)

	return renderer.NewChainRenderer(ctx.renderer).
		Render(describeDeploymentsReply).
		Render(deploymentDefinitionReply).
		Render(listRegionalDeploymentsReply).
		Render(listInstancesReply).
		Err()
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
	return []string{"id", "service", "status", "messages", "regions", "created_at", "updated_at", "definition"}
}

func (r *DescribeDeploymentReply) Fields() []map[string]string {
	item := r.value.GetDeployment()
	fields := map[string]string{
		"id":         renderer.FormatDeploymentID(r.mapper, item.GetId(), r.full),
		"service":    renderer.FormatServiceSlug(r.mapper, item.GetServiceId(), r.full),
		"status":     formatDeploymentStatus(item.GetStatus()),
		"messages":   formatDeploymentMessages(item.GetMessages(), 0),
		"regions":    renderRegions(item.Definition.Regions),
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
		"updated_at": renderer.FormatTime(item.GetUpdatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}

// DescribeDeploymentDefinitionReply implements resources.ApiResources to display the deployment definition as YAML.
type DescribeDeploymentDefinitionReply struct {
	value *koyeb.GetDeploymentReply
}

func NewDescribeDeploymentDefinitionReply(value *koyeb.GetDeploymentReply) *DescribeDeploymentDefinitionReply {
	return &DescribeDeploymentDefinitionReply{
		value: value,
	}
}

func (DescribeDeploymentDefinitionReply) Title() string {
	return "Definition"
}

func (r *DescribeDeploymentDefinitionReply) MarshalBinary() ([]byte, error) {
	return r.value.GetDeployment().Definition.MarshalJSON()
}

func (r *DescribeDeploymentDefinitionReply) Headers() []string {
	return []string{"content"}
}

func (r *DescribeDeploymentDefinitionReply) Fields() []map[string]string {
	item := r.value.GetDeployment()

	json, err := item.Definition.MarshalJSON()
	if err != nil {
		return nil
	}
	b, err := yaml.JSONToYAML(json)
	if err != nil {
		return nil
	}

	fields := map[string]string{
		"content": string(b),
	}

	return []map[string]string{fields}
}
