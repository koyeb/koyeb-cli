package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

func (h *InstanceHandler) Log(cmd *cobra.Command, args []string) error {
	instanceDetail, _, err := h.client.InstancesApi.GetInstance(h.ctx, h.ResolveInstanceArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	done := make(chan struct{})

	query := &watchLogQuery{instanceID: koyeb.PtrString(instanceDetail.Instance.GetId())}
	return watchLog(query, done)
}
