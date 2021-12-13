package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) Describe(cmd *cobra.Command, args []string) error {
	res, _, err := h.client.SecretsApi.GetSecret(h.ctxWithAuth, h.ResolveSecretArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getSecretsReply := NewDescribeSecretReply(h.mapper, &res, full)

	return renderer.NewDescribeRenderer(getSecretsReply).Render(output)

}

type DescribeSecretReply struct {
	mapper *idmapper.Mapper
	res    *koyeb.GetSecretReply
	full   bool
}

func NewDescribeSecretReply(mapper *idmapper.Mapper, res *koyeb.GetSecretReply, full bool) *DescribeSecretReply {
	return &DescribeSecretReply{
		mapper: mapper,
		res:    res,
		full:   full,
	}
}

func (a *DescribeSecretReply) MarshalBinary() ([]byte, error) {
	return a.res.GetSecret().MarshalJSON()
}

func (a *DescribeSecretReply) Title() string {
	return "Secret"
}

func (a *DescribeSecretReply) Headers() []string {
	return []string{"id", "name", "type", "value", "created_at", "updated_at"}
}

func (a *DescribeSecretReply) Fields() []map[string]string {
	res := []map[string]string{}
	item := a.res.GetSecret()
	fields := map[string]string{
		"id":         renderer.FormatSecretID(a.mapper, item.GetId(), a.full),
		"name":       item.GetName(),
		"type":       formatSecretType(item.GetType()),
		"value":      "*****",
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
		"updated_at": renderer.FormatTime(item.GetUpdatedAt()),
	}
	res = append(res, fields)
	return res
}
