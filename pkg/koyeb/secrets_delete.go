package koyeb

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) Delete(cmd *cobra.Command, args []string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	_, _, err := client.SecretsApi.DeleteSecret(ctx, ResolveSecretShortID(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}
	log.Infof("Secret %s deleted.", args[0])
	return nil
}
