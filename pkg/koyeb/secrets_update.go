package koyeb

import (
	"bufio"
	"os"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) Update(ctx *CLIContext, cmd *cobra.Command, args []string, updateSecret *koyeb.Secret) error {
	if cmd.LocalFlags().Lookup("value-from-stdin").Changed && cmd.LocalFlags().Lookup("value").Changed {
		log.Fatalf("Cannot use value and value-from-stdin at the same time")
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
	res, resp, err := ctx.client.SecretsApi.UpdateSecret2(ctx.context, ResolveSecretArgs(ctx, args[0])).Secret(*updateSecret).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	getSecretsReply := NewGetSecretReply(ctx.mapper, &koyeb.GetSecretReply{Secret: res.Secret}, full)
	return ctx.renderer.Render(getSecretsReply)
}
