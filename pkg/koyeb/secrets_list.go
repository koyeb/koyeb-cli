package koyeb

import (
	"context"
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) List(cmd *cobra.Command, args []string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())
	results := koyeb.ListSecretsReply{}

	page := 0
	offset := 0
	limit := 100
	for {
		res, _, err := client.SecretsApi.ListSecrets(ctx).Limit(fmt.Sprintf("%d", limit)).Offset(fmt.Sprintf("%d", offset)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		if results.Secrets == nil {
			results = res
		} else {
			*results.Secrets = append(*results.Secrets, *res.Secrets...)
		}
		page += 1
		offset = page * limit
		if int64(offset) >= res.GetCount() {
			break
		}
	}

	full, _ := cmd.Flags().GetBool("full")
	listSecretsReply := NewListSecretsReply(&results, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.NewListRenderer(listSecretsReply).Render(output)
}

type ListSecretsReply struct {
	res  *koyeb.ListSecretsReply
	full bool
}

func NewListSecretsReply(res *koyeb.ListSecretsReply, full bool) *ListSecretsReply {
	return &ListSecretsReply{
		res:  res,
		full: full,
	}
}

func (a *ListSecretsReply) Title() string {
	return "Secrets"
}

func (a *ListSecretsReply) MarshalBinary() ([]byte, error) {
	return a.res.MarshalJSON()
}

func (a *ListSecretsReply) Headers() []string {
	return []string{"id", "name", "type", "value", "created_at"}
}

func (a *ListSecretsReply) Fields() []map[string]string {
	res := []map[string]string{}
	for _, item := range a.res.GetSecrets() {
		fields := map[string]string{
			"id":         renderer.FormatID(item.GetId(), a.full),
			"name":       item.GetName(),
			"type":       formatSecretType(item.GetType()),
			"value":      "*****",
			"created_at": renderer.FormatTime(item.GetCreatedAt()),
		}
		res = append(res, fields)
	}
	return res
}
