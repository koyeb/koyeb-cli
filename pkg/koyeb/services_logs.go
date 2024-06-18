package koyeb

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

func (h *ServiceHandler) Logs(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	serviceName, err := h.parseServiceName(cmd, args[0])
	if err != nil {
		return err
	}
	service, err := h.ResolveServiceArgs(ctx, serviceName)
	if err != nil {
		return err
	}

	serviceDetail, resp, err := ctx.Client.ServicesApi.GetService(ctx.Context, service).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the service `%s`", serviceName),
			err,
			resp,
		)
	}

	logsType := GetStringFlags(cmd, "type")
	serviceId := serviceDetail.Service.GetId()
	deploymentId := ""
	instanceId := GetStringFlags(cmd, "instance")

	latestDeployList, resp, err := ctx.Client.DeploymentsApi.ListDeployments(ctx.Context).
		Limit("1").ServiceId(service).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while listing the deployments of the service `%s`", serviceName),
			err,
			resp,
		)
	}
	if len(latestDeployList.GetDeployments()) == 0 {
		return &errors.CLIError{
			What: "Error while fetching the logs of your service",
			Why:  "we couldn't find the latest deployment of your service",
			Additional: []string{
				"Your service exists but has not been deployed yet",
			},
			Orig:     nil,
			Solution: "Try again in a few seconds. If the problem persists, please create an issue on https://github.com/koyeb/koyeb-cli/issues/new",
		}
	}

	latestDeploymentId := *latestDeployList.GetDeployments()[0].Id

	reply, _, err := ctx.Client.DeploymentsApi.GetDeployment(ctx.Context, latestDeploymentId).Execute()
	if err != nil {
		return &errors.CLIError{
			What: "Error while fetching the logs of your service",
			Why:  "we couldn't find the latest deployment of your service",
			Additional: []string{
				"Your service is nowhere to be found in our systems",
			},
			Orig:     nil,
			Solution: "Please contact us.",
		}
	}
	if reply.Deployment.SkipBuild != nil && *reply.Deployment.SkipBuild {
		if serviceDetail.Service.LastProvisionedDeploymentId == nil {
			return &errors.CLIError{
				What: "Error while fetching the logs of your service",
				Why:  "we couldn't find the latest provisioned deployment of your service",
				Additional: []string{
					"Your service is nowhere to be found in our systems",
				},
				Orig:     nil,
				Solution: "Please contact us.",
			}
		}
		logrus.Warnf("This deployment uses a previous build originally created during deployment %s. If you want to access those, use `koyeb deployments logs -t build %s`", *serviceDetail.Service.LastProvisionedDeploymentId, *serviceDetail.Service.LastProvisionedDeploymentId)
		if logsType == "build" {
			return nil
		}
	}

	if logsType == "build" {
		deploymentId = latestDeploymentId
	}

	logsQuery, err := ctx.LogsClient.NewWatchLogsQuery(
		logsType,
		serviceId,
		deploymentId,
		instanceId,
		GetBoolFlags(cmd, "full"),
	)
	if err != nil {
		return err
	}
	return logsQuery.PrintAll()
}
