package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper2"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) Get(cmd *cobra.Command, args []string) error {
	res, _, err := h.client.SecretsApi.GetSecret(h.ctxWithAuth, h.ResolveSecretArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getSecretsReply := NewGetSecretReply(h.mapper, &res, full)

	return renderer.NewItemRenderer(getSecretsReply).Render(output)

}

type GetSecretReply struct {
	mapper *idmapper2.Mapper
	res    *koyeb.GetSecretReply
	full   bool
}

func NewGetSecretReply(mapper *idmapper2.Mapper, res *koyeb.GetSecretReply, full bool) *GetSecretReply {
	return &GetSecretReply{
		mapper: mapper,
		res:    res,
		full:   full,
	}
}

func (a *GetSecretReply) MarshalBinary() ([]byte, error) {
	return a.res.GetSecret().MarshalJSON()
}

func (a *GetSecretReply) Title() string {
	return "Secret"
}

func (a *GetSecretReply) Headers() []string {
	return []string{"id", "name", "type", "value", "created_at"}
}

func (a *GetSecretReply) Fields() []map[string]string {
	res := []map[string]string{}
	item := a.res.GetSecret()
	fields := map[string]string{
		"id":         renderer.FormatSecretID(a.mapper, item.GetId(), a.full),
		"name":       item.GetName(),
		"type":       formatSecretType(item.GetType()),
		"value":      "*****",
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
	}
	res = append(res, fields)
	return res
}

func formatSecretType(st koyeb.SecretType) string {
	return fmt.Sprintf("%s", st)
}
