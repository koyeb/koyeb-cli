package koyeb

import (
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) List(cmd *cobra.Command, args []string) error {
	list := []koyeb.Secret{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		res, _, err := h.client.SecretsApi.ListSecrets(h.ctxWithAuth).
			Limit(strconv.FormatInt(limit, 10)).Offset(strconv.FormatInt(offset, 10)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		list = append(list, res.GetSecrets()...)

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	listSecretsReply := NewListSecretsReply(h.mapper, &koyeb.ListSecretsReply{Secrets: &list}, full)

	return renderer.NewListRenderer(listSecretsReply).Render(output)
}

type ListSecretsReply struct {
	mapper *idmapper.Mapper
	res    *koyeb.ListSecretsReply
	full   bool
}

func NewListSecretsReply(mapper *idmapper.Mapper, res *koyeb.ListSecretsReply, full bool) *ListSecretsReply {
	return &ListSecretsReply{
		mapper: mapper,
		res:    res,
		full:   full,
	}
}

func (ListSecretsReply) Title() string {
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
			"id":         renderer.FormatSecretID(a.mapper, item.GetId(), a.full),
			"name":       item.GetName(),
			"type":       formatSecretType(item.GetType()),
			"value":      "*****",
			"created_at": renderer.FormatTime(item.GetCreatedAt()),
		}
		res = append(res, fields)
	}

	return res
}
