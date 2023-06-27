package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) Describe(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	secret, err := ResolveSecretArgs(ctx, args[0])
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.SecretsApi.GetSecret(ctx.Context, secret).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the secret `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getSecretsReply := NewDescribeSecretReply(ctx.Mapper, res, full)
	ctx.Renderer.Render(getSecretsReply)
	return nil
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
