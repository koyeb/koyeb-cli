package koyeb

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/dates"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

func (h *ServiceHandler) Logs(ctx *CLIContext, cmd *cobra.Command, since time.Time, args []string) error {
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

	startStr := GetStringFlags(cmd, "start-time")
	endStr := GetStringFlags(cmd, "end-time")
	regex := GetStringFlags(cmd, "regex-search")
	text := GetStringFlags(cmd, "text-search")
	order := GetStringFlags(cmd, "order")

	if !since.IsZero() && startStr != "" {
		return &errors.CLIError{
			What: "Error while fetching logs",
			Why:  "Cannot use since with start-time",
		}
	}

	end := time.Now()
	if endStr != "" {
		end, err = dates.Parse(endStr)
		if err != nil {
			return &errors.CLIError{
				What:     "Error while fetching logs",
				Why:      "End time was improperly formatted.",
				Orig:     err,
				Solution: "Enter end time using this layout: '2006-01-02 15:04:05'",
			}
		}
	}
	start := end.Add(-5 * time.Minute)
	if !since.IsZero() {
		start = since
	}
	if startStr != "" {
		start, err = dates.Parse(startStr)
		if err != nil {
			return &errors.CLIError{
				What:     "Error while fetching logs",
				Why:      "start time was improperly formatted.",
				Orig:     err,
				Solution: "Enter start time using this layout: '2006-01-02 15:04:05'",
			}
		}
	}

	err = h.queryLogs(ctx, logsType, serviceId, deploymentId, instanceId, start, end, regex, text, order, GetBoolFlags(cmd, "full"))
	if err != nil {
		return err
	}

	if endStr != "" {
		return nil
	}
	logsQuery, err := ctx.LogsClient.NewWatchLogsQuery(
		logsType,
		serviceId,
		deploymentId,
		instanceId,
		end,
		GetBoolFlags(cmd, "full"),
	)
	if err != nil {
		return err
	}
	return logsQuery.PrintAll()
}

func (h *ServiceHandler) queryLogs(ctx *CLIContext, logsType, serviceId, deploymentId, instanceId string, start, end time.Time, regex, text, order string, full bool) error {
	hasMore := true

	for hasMore {
		resp, err := ctx.LogsClient.ExecuteQueryLogsQuery(
			ctx.Context,
			logsType,
			serviceId,
			deploymentId,
			instanceId,
			start,
			end,
			regex,
			text,
			order,
		)
		if err != nil {
			return err
		}

		if resp.Pagination.HasMore != nil {
			hasMore = *resp.Pagination.HasMore
			if hasMore {
				start = *resp.Pagination.NextStart
				end = *resp.Pagination.NextEnd
			}
		}

		for _, log := range resp.Data {
			stream, ok := log.Labels["stream"].(string)
			if !ok {
				stream = ""
			}
			instance_id, ok := log.Labels["instance_id"].(string)
			if !ok {
				instance_id = ""
			}
			err := PrintLogLine(log, full, *log.CreatedAt, *log.Msg, stream, instance_id)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
