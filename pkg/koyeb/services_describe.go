package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Describe(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	serviceName, err := h.parseServiceName(cmd, args[0])
	if err != nil {
		return err
	}

	service, err := h.ResolveServiceArgs(ctx, serviceName)
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.ServicesApi.GetService(ctx.Context, service).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the service `%s`", serviceName),
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
			fmt.Sprintf("Error while listing the instances of the service `%s`", serviceName),
			err,
			resp,
		)
	}

	deploymentsRes, resp, err := ctx.Client.DeploymentsApi.ListDeployments(ctx.Context).ServiceId(res.Service.GetId()).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while listing the deployments of the service `%s`", serviceName),
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
