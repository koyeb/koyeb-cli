package koyeb

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

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

	// if either start or end are provided,
	// query logs, then return
	if startStr != "" || endStr != "" {
		end := time.Now()
		if endStr != "" {
			layout := "2006-01-02 15:04:05 +0000 UTC"
			end, err = time.Parse(layout, endStr)
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
		if startStr != "" {
			layout := "2006-01-02 15:04:05 +0000 UTC"
			start, err = time.Parse(layout, startStr)
			if err != nil {
				return &errors.CLIError{
					What:     "Error while fetching logs",
					Why:      "start time was improperly formatted.",
					Orig:     err,
					Solution: "Enter start time using this layout: '2006-01-02 15:04:05'",
				}
			}
		}

		prevLogs, err := ctx.LogsClient.ExecuteQueryLogsQuery(
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
			GetBoolFlags(cmd, "full"),
		)
		if err != nil {
			return err
		}
		for _, log := range prevLogs {
			err := PrintLogLine(log, GetBoolFlags(cmd, "full"), *log.CreatedAt, *log.Msg, log.Labels["stream"].(string), log.Labels["instance_id"].(string))
			if err != nil {
				return err
			}
		}

		return nil
	}

	if since.IsZero() {
		since = serviceDetail.Service.GetCreatedAt()
	}

	logsQuery, err := ctx.LogsClient.NewWatchLogsQuery(
		logsType,
		serviceId,
		deploymentId,
		instanceId,
		since,
		GetBoolFlags(cmd, "full"),
	)
	if err != nil {
		return err
	}
	return logsQuery.PrintAll()
}
