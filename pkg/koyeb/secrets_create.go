package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string, createSecret *koyeb.CreateSecret) error {
	res, resp, err := ctx.Client.SecretsApi.CreateSecret(ctx.Context).Secret(*createSecret).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while creating the secret `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getSecretsReply := NewGetSecretReply(ctx.Mapper, &koyeb.GetSecretReply{Secret: res.Secret}, full)
	ctx.Renderer.Render(getSecretsReply)
	return nil
}
