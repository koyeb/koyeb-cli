package koyeb

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string, createSecret *koyeb.CreateSecret) error {
	createSecret.SetName(args[0])
	if !cmd.LocalFlags().Lookup("value-from-stdin").Changed && !cmd.LocalFlags().Lookup("value").Changed {
		prompt := promptui.Prompt{
			Label: "Enter your secret",
			Mask:  '*',
		}

		result, err := prompt.Run()
		// When user cancels the prompt, we return nil to cancel the command
		if err != nil {
			return nil
		}
		createSecret.SetValue(result)
	}
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

		createSecret.SetValue(strings.Join(input, "\n"))
	}
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
