package koyeb

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) Update(ctx *CLIContext, cmd *cobra.Command, args []string, updateSecret *koyeb.Secret) error {
	if cmd.LocalFlags().Lookup("value-from-stdin").Changed && cmd.LocalFlags().Lookup("value").Changed {
		return &errors.CLIError{
			What:       "Invalid arguments to create a secret",
			Why:        "you can't provide both --value and --value-from-stdin at the same time",
			Additional: nil,
			Orig:       nil,
			Solution:   "Remove one of the flags",
		}
	}
	if cmd.LocalFlags().Lookup("value-from-stdin").Changed {
		var input []string

		scanner := bufio.NewScanner(os.Stdin)
		for {
			scanner.Scan()
			text := scanner.Text()
			if len(text) != 0 {
				input = append(input, text)
			} else {
				break
			}
		}
		updateSecret.SetValue(strings.Join(input, "\n"))
	}

	secret, err := ResolveSecretArgs(ctx, args[0])
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.SecretsApi.UpdateSecret2(ctx.Context, secret).Secret(*updateSecret).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while updating the secret `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getSecretsReply := NewGetSecretReply(ctx.Mapper, &koyeb.GetSecretReply{Secret: res.Secret}, full)
	ctx.Renderer.Render(getSecretsReply)
	return nil
}
