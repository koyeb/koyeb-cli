package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

func (h *InstanceHandler) Logs(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	instanceDetail, resp, err := ctx.Client.InstancesApi.GetInstance(ctx.Context, h.ResolveInstanceArgs(ctx, args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	done := make(chan struct{})

	query := &WatchLogQuery{}
	query.InstanceID = koyeb.PtrString(instanceDetail.Instance.GetId())

	return WatchLog(query, done)
}
