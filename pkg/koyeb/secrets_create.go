package koyeb

import (
	"bufio"
	"os"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) Create(cmd *cobra.Command, args []string, createSecret *koyeb.CreateSecret) error {
	createSecret.SetName(args[0])
	if !cmd.LocalFlags().Lookup("value-from-stdin").Changed && !cmd.LocalFlags().Lookup("value").Changed {
		prompt := promptui.Prompt{
			Label: "Enter your secret",
			Mask:  '*',
		}

		result, err := prompt.Run()
		if err != nil {
			er(err)
		}
		createSecret.SetValue(result)
	}
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

		createSecret.SetValue(strings.Join(input, "\n"))
	}
	res, _, err := h.client.SecretsApi.CreateSecret(h.ctx).Body(*createSecret).Execute()
	if err != nil {
		fatalApiError(err)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getSecretsReply := NewGetSecretReply(h.mapper, &koyeb.GetSecretReply{Secret: res.Secret}, full)

	return renderer.NewDescribeItemRenderer(getSecretsReply).Render(output)
}
