package koyeb

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper2"
	"github.com/spf13/cobra"
)

func NewSecretCmd() *cobra.Command {
	h := NewSecretHandler()

	secretCmd := &cobra.Command{
		Use:               "secrets ACTION",
		Aliases:           []string{"sec", "secret"},
		Short:             "Secrets",
		PersistentPreRunE: h.InitHandler,
	}

	createSecretCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			createSecret := koyeb.NewCreateSecretWithDefaults()
			SyncFlags(cmd, args, createSecret)
			return h.Create(cmd, args, createSecret)
		},
	}
	createSecretCmd.Flags().StringP("value", "v", "", "Secret Value")
	createSecretCmd.Flags().Bool("value-from-stdin", false, "Secret Value from stdin")
	secretCmd.AddCommand(createSecretCmd)

	getSecretCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get secret",
		Args:  cobra.ExactArgs(1),
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
		Use:   "describe NAME",
		Short: "Describe secret",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Describe,
	}
	secretCmd.AddCommand(describeSecretCmd)

	updateSecretCmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			updateSecret := koyeb.NewSecretWithDefaults()
			SyncFlags(cmd, args, updateSecret)
			return h.Update(cmd, args, updateSecret)
		},
	}
	updateSecretCmd.Flags().StringP("value", "v", "", "Secret Value")
	updateSecretCmd.Flags().Bool("value-from-stdin", false, "Secret Value from stdin")
	secretCmd.AddCommand(updateSecretCmd)

	deleteSecretCmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete secret",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Delete,
	}
	secretCmd.AddCommand(deleteSecretCmd)

	return secretCmd
}

func NewSecretHandler() *SecretHandler {
	return &SecretHandler{}
}

type SecretHandler struct {
	ctxWithAuth context.Context
	client      *koyeb.APIClient
	mapper      *idmapper2.Mapper
}

func (h *SecretHandler) InitHandler(cmd *cobra.Command, args []string) error {
	h.client = getApiClient()
	h.ctxWithAuth = getAuth(context.Background())
	h.mapper = idmapper2.NewMapper(h.ctxWithAuth, h.client)
	return nil
}

func (h *SecretHandler) ResolveSecretArgs(val string) string {
	secretMapper := h.mapper.Secret()
	id, err := secretMapper.ResolveID(val)
	if err != nil {
		fatalApiError(err)
	}

	return id
}
