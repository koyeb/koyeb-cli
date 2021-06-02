package koyeb

import (
	"context"
	"fmt"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

func NewSecretCmd() *cobra.Command {
	h := NewSecretHandler()

	secretCmd := &cobra.Command{
		Use:     "secrets [action]",
		Aliases: []string{"s", "secret"},
		Short:   "Secrets",
	}

	createSecretCmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create secrets",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			createSecret := koyeb.NewCreateSecretWithDefaults()
			SyncFlags(cmd, args, createSecret)
			return h.Create(cmd, args, createSecret)
		},
	}
	createSecretCmd.Flags().StringP("value", "v", "", "Secret Value")
	secretCmd.AddCommand(createSecretCmd)

	getSecretCmd := &cobra.Command{
		Use:   "get [name]",
		Short: "Get secret",
		RunE:  h.Get,
	}
	secretCmd.AddCommand(getSecretCmd)

	listSecretCmd := &cobra.Command{
		Use:   "list",
		Short: "List secrets",
		RunE:  h.List,
	}
	secretCmd.AddCommand(listSecretCmd)

	describeSecretCmd := &cobra.Command{
		Use:   "describe [name]",
		Short: "Describe secrets",
		RunE:  h.Describe,
	}
	secretCmd.AddCommand(describeSecretCmd)

	deleteSecretCmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete secrets",
		Args:  cobra.MinimumNArgs(1),
		RunE:  h.Delete,
	}
	secretCmd.AddCommand(deleteSecretCmd)

	return secretCmd
}

func NewSecretHandler() *SecretHandler {
	return &SecretHandler{}
}

type SecretHandler struct {
}

func (h *SecretHandler) Create(cmd *cobra.Command, args []string, createSecret *koyeb.CreateSecret) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	createSecret.SetName(args[0])
	_, _, err := client.SecretsApi.CreateSecret(ctx).Body(*createSecret).Execute()
	if err != nil {
		fatalApiError(err)
	}
	return nil
}

func (h *SecretHandler) Get(cmd *cobra.Command, args []string) error {
	format := "table"
	if len(args) == 0 {
		return h.listFormat(cmd, args, format)
	}
	return h.getFormat(cmd, args, format)
}

func (h *SecretHandler) Describe(cmd *cobra.Command, args []string) error {
	format := "yaml"
	if len(args) == 0 {
		return h.listFormat(cmd, args, format)
	}
	return h.getFormat(cmd, args, format)
}

func (h *SecretHandler) Delete(cmd *cobra.Command, args []string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	for _, arg := range args {
		_, _, err := client.SecretsApi.DeleteSecret(ctx, arg).Execute()
		if err != nil {
			fatalApiError(err)
		}
	}
	return nil
}

func (h *SecretHandler) List(cmd *cobra.Command, args []string) error {
	format := "table"
	return h.listFormat(cmd, args, format)
}

func (h *SecretHandler) getFormat(cmd *cobra.Command, args []string, format string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	var items []ApiResources
	for _, arg := range args {
		res, _, err := client.SecretsApi.GetSecret(ctx, arg).Execute()
		if err != nil {
			fatalApiError(err)
		}
		items = append(items, &GetSecretReply{res})
	}

	render(format, items...)

	return nil
}

func (h *SecretHandler) listFormat(cmd *cobra.Command, args []string, format string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	var items []ApiResources

	page := 0
	offset := 0
	limit := 10
	for {
		res, _, err := client.SecretsApi.ListSecrets(ctx).Limit(fmt.Sprintf("%d", limit)).Offset(fmt.Sprintf("%d", offset)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		items = append(items, &ListSecretsReply{res})
		page += 1
		offset = page * limit
		if int64(offset) >= res.GetCount() {
			break
		}
	}

	render(format, items...)

	return nil
}

type GetSecretReply struct {
	koyeb.GetSecretReply
}

func (a *GetSecretReply) MarshalBinary() ([]byte, error) {
	return a.GetSecretReply.GetSecret().MarshalJSON()
}

func (a *GetSecretReply) GetTableHeaders() []string {
	return []string{"id", "name", "value", "updated_at"}
}

func (a *GetSecretReply) GetTableValues() [][]string {
	var res [][]string
	item := a.GetSecret()
	var fields []string
	for _, field := range a.GetTableHeaders() {
		fields = append(fields, GetField(item, field))
	}
	res = append(res, fields)
	return res
}

type ListSecretsReply struct {
	koyeb.ListSecretsReply
}

func (a *ListSecretsReply) MarshalBinary() ([]byte, error) {
	return a.ListSecretsReply.MarshalJSON()
}

func (a *ListSecretsReply) GetTableHeaders() []string {
	return []string{"id", "name", "value", "updated_at"}
}

func (a *ListSecretsReply) GetTableValues() [][]string {
	var res [][]string
	for _, item := range a.GetSecrets() {
		var fields []string
		for _, field := range a.GetTableHeaders() {
			fields = append(fields, GetField(item, field))
		}
		res = append(res, fields)
	}
	return res
}
