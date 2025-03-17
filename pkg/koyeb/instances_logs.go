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

	startStr := GetStringFlags(cmd, "start-time")
	endStr := GetStringFlags(cmd, "end-time")
	regex := GetStringFlags(cmd, "regex-search")
	text := GetStringFlags(cmd, "text-search")
	order := GetStringFlags(cmd, "order")
	tail := GetBoolFlags(cmd, "tail")
	output := GetStringFlags(cmd, "output")

	return ctx.LogsClient.PrintLogs(ctx, LogsQuery{
		InstanceId: instanceDetail.Instance.GetId(),
		Start:      startStr,
		End:        endStr,
		Since:      since,
		Regex:      regex,
		Text:       text,
		Order:      order,
		Tail:       tail,
		Output:     output,
		Full:       GetBoolFlags(cmd, "full"),
	})
}
