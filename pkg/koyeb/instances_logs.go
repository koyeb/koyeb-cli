package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

func (h *InstanceHandler) Logs(cmd *cobra.Command, args []string) error {
	instanceDetail, _, err := h.client.InstancesApi.GetInstance(h.ctx, h.ResolveInstanceArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	done := make(chan struct{})

	query := &WatchLogQuery{}
	query.InstanceID = koyeb.PtrString(instanceDetail.Instance.GetId())

	return WatchLog(query, done)
}
