package koyeb

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Delete(cmd *cobra.Command, args []string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	_, _, err := client.ServicesApi.DeleteService(ctx, ResolveServiceShortID(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}
	log.Infof("Service %s deleted.", args[0])
	return nil
}
