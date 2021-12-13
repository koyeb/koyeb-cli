package koyeb

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) Delete(cmd *cobra.Command, args []string) error {
	_, _, err := h.client.SecretsApi.DeleteSecret(h.ctx, h.ResolveSecretArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}
	log.Infof("Secret %s deleted.", args[0])
	return nil
}
