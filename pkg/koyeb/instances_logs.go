package koyeb

import (
	"fmt"
	"time"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *InstanceHandler) Logs(ctx *CLIContext, cmd *cobra.Command, since time.Time, args []string) error {
	instance, err := h.ResolveInstanceArgs(ctx, args[0])
	if err != nil {
		return err
	}

	instanceDetail, resp, err := ctx.Client.InstancesApi.GetInstance(ctx.Context, instance).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the logs of the instance `%s`", args[0]),
			err,
			resp,
		)
	}
	logsQuery, err := ctx.LogsClient.NewWatchLogsQuery(
		"",
		"",
		"",
		instanceDetail.Instance.GetId(),
		since,
		GetBoolFlags(cmd, "full"),
	)
	if err != nil {
		return err
	}
	return logsQuery.PrintAll()
}
