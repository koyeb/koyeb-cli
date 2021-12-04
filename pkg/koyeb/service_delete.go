package koyeb

import (
	"context"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Delete(cmd *cobra.Command, args []string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	for _, arg := range args {
		_, _, err := client.ServicesApi.DeleteService(ctx, ResolveServiceShortID(arg)).Execute()
		if err != nil {
			fatalApiError(err)
		}
	}
	log.Infof("Services %s deleted.", strings.Join(args, ", "))
	return nil
}
