package koyeb

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) Delete(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	_, resp, err := ctx.Client.SecretsApi.DeleteSecret(ctx.Context, ResolveSecretArgs(ctx, args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}
	log.Infof("Secret %s deleted.", args[0])
	return nil
}
