package koyeb

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

func (h *InstanceHandler) Log(cmd *cobra.Command, args []string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	instanceDetail, _, err := client.InstancesApi.GetInstance(ctx, ResolveInstanceShortID(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	done := make(chan struct{})

	query := &watchLogQuery{instanceID: koyeb.PtrString(instanceDetail.Instance.GetId())}
	return watchLog(query, done)
}
