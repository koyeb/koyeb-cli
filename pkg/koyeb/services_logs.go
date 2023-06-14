package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Logs(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	serviceDetail, resp, err := ctx.client.ServicesApi.GetService(ctx.context, h.ResolveServiceArgs(ctx, args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	done := make(chan struct{})

	serviceID := serviceDetail.Service.GetId()
	logType := GetStringFlags(cmd, "type")
	instanceID := GetStringFlags(cmd, "instance")

	query := &WatchLogQuery{}
	query.ServiceID = koyeb.PtrString(serviceID)

	if logType == "build" {
		latestDeploy, resp, err := ctx.client.DeploymentsApi.ListDeployments(ctx.context).
			Limit("1").ServiceId(h.ResolveServiceArgs(ctx, args[0])).Execute()
		if err != nil {
			fatalApiError(err, resp)
		}
		if len(latestDeploy.GetDeployments()) == 0 {
			return errors.New("unable to load latest deployment")
		}
		query.DeploymentID = latestDeploy.GetDeployments()[0].Id
		query.LogType = koyeb.PtrString(logType)
	}

	if instanceID != "" {
		query.InstanceID = koyeb.PtrString(instanceID)
	}

	return WatchLog(query, done)
}
