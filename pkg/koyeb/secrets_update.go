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
	res, _, err := h.client.SecretsApi.UpdateSecret2(h.ctxWithAuth, h.ResolveSecretShortID(args[0])).Body(*updateSecret).Execute()
	if err != nil {
		fatalApiError(err)
	}
	full, _ := cmd.Flags().GetBool("full")
	getSecretsReply := NewGetSecretReply(&koyeb.GetSecretReply{Secret: res.Secret}, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.NewDescribeItemRenderer(getSecretsReply).Render(output)
}
