package koyeb

import (
	"context"
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) Get(cmd *cobra.Command, args []string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	res, _, err := client.SecretsApi.GetSecret(ctx, ResolveSecretShortID(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}
	full, _ := cmd.Flags().GetBool("full")
	getSecretsReply := NewGetSecretReply(&res, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.NewItemRenderer(getSecretsReply).Render(output)

}

type GetSecretReply struct {
	res  *koyeb.GetSecretReply
	full bool
}

func NewGetSecretReply(res *koyeb.GetSecretReply, full bool) *GetSecretReply {
	return &GetSecretReply{
		res:  res,
		full: full,
	}
}

func (a *GetSecretReply) MarshalBinary() ([]byte, error) {
	return a.res.GetSecret().MarshalJSON()
}

func (a *GetSecretReply) Title() string {
	return "Secret"
}

func (a *GetSecretReply) Headers() []string {
	return []string{"id", "name", "type", "value", "updated_at"}
}

func (a *GetSecretReply) Fields() []map[string]string {
	res := []map[string]string{}
	item := a.res.GetSecret()
	fields := map[string]string{
		"id":         renderer.FormatID(item.GetId(), a.full),
		"name":       item.GetName(),
		"type":       formatSecretType(item.GetType()),
		"value":      "*****",
		"updated_at": renderer.FormatTime(item.GetUpdatedAt()),
	}
	res = append(res, fields)
	return res
}

func formatSecretType(st koyeb.SecretType) string {
	return fmt.Sprintf("%s", st)
}
