package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) Delete(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	secret, err := ResolveSecretArgs(ctx, args[0])
	if err != nil {
		return err
	}

	_, resp, err := ctx.Client.SecretsApi.DeleteSecret(ctx.Context, secret).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while deleting the secret `%s`", args[0]),
			err,
			resp,
		)
	}
	log.Infof("Secret %s deleted.", args[0])
	return nil
}
