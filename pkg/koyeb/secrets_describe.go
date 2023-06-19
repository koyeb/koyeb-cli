package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) Describe(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	res, resp, err := ctx.client.SecretsApi.GetSecret(ctx.context, ResolveSecretArgs(ctx, args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	getSecretsReply := NewDescribeSecretReply(ctx.mapper, res, full)
	return ctx.renderer.Render(getSecretsReply)

}

type DescribeSecretReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetSecretReply
	full   bool
}

func NewDescribeSecretReply(mapper *idmapper.Mapper, value *koyeb.GetSecretReply, full bool) *DescribeSecretReply {
	return &DescribeSecretReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (DescribeSecretReply) Title() string {
	return "Secret"
}

func (r *DescribeSecretReply) MarshalBinary() ([]byte, error) {
	return r.value.GetSecret().MarshalJSON()
}

func (r *DescribeSecretReply) Headers() []string {
	return []string{"id", "name", "type", "value", "created_at", "updated_at"}
}

func (r *DescribeSecretReply) Fields() []map[string]string {
	item := r.value.GetSecret()
	fields := map[string]string{
		"id":         renderer.FormatSecretID(r.mapper, item.GetId(), r.full),
		"name":       item.GetName(),
		"type":       formatSecretType(item.GetType()),
		"value":      "*****",
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
		"updated_at": renderer.FormatTime(item.GetUpdatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}
