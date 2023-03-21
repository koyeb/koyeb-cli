package koyeb

import (
	"bufio"
	"os"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) Update(cmd *cobra.Command, args []string, updateSecret *koyeb.Secret) error {
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
	res, resp, err := h.client.SecretsApi.UpdateSecret2(h.ctx, h.ResolveSecretArgs(args[0])).Secret(*updateSecret).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getSecretsReply := NewGetSecretReply(h.mapper, &koyeb.GetSecretReply{Secret: res.Secret}, full)

	return renderer.NewDescribeItemRenderer(getSecretsReply).Render(output)
}
