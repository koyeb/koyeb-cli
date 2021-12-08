package koyeb

import (
	"context"

	"github.com/spf13/cobra"
)

func (h *SecretHandler) Delete(cmd *cobra.Command, args []string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	for _, arg := range args {
		_, _, err := client.SecretsApi.DeleteSecret(ctx, arg).Execute()
		if err != nil {
			fatalApiError(err)
		}
	}
	return nil
}
