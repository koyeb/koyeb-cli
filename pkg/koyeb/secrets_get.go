package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *SecretHandler) Get(ctx *CLIContext, cmd *cobra.Command, args []string) error {
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
	getSecretsReply := NewGetSecretReply(ctx.Mapper, res, full)
	ctx.Renderer.Render(getSecretsReply)
	return nil
}

type GetSecretReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetSecretReply
	full   bool
}

func NewGetSecretReply(mapper *idmapper.Mapper, res *koyeb.GetSecretReply, full bool) *GetSecretReply {
	return &GetSecretReply{
		mapper: mapper,
		value:  res,
		full:   full,
	}
}

func (GetSecretReply) Title() string {
	return "Secret"
}

func (r *GetSecretReply) MarshalBinary() ([]byte, error) {
	return r.value.GetSecret().MarshalJSON()
}

func (r *GetSecretReply) Headers() []string {
	return []string{"id", "name", "type", "value", "created_at"}
}

func (r *GetSecretReply) Fields() []map[string]string {
	item := r.value.GetSecret()
	fields := map[string]string{
		"id":         renderer.FormatID(item.GetId(), r.full),
		"name":       item.GetName(),
		"type":       formatSecretType(item.GetType()),
		"value":      "*****",
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}

func formatSecretType(st koyeb.SecretType) string {
	return string(st)
}
