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
		Use:     "secrets ACTION",
		Aliases: []string{"sec", "secret"},
		Short:   "Secrets",
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
}

func buildSecretShortIDCache() map[string][]string {
	c := make(map[string][]string)
	client := getApiClient()
	ctx := getAuth(context.Background())

	page := 0
	offset := 0
	limit := 100
	for {
		res, _, err := client.SecretsApi.ListSecrets(ctx).Limit(fmt.Sprintf("%d", limit)).Offset(fmt.Sprintf("%d", offset)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		for _, a := range *res.Secrets {
			id := a.GetId()[:8]
			c[id] = append(c[id], a.GetId())

		}

		page += 1
		offset = page * limit
		if int64(offset) >= res.GetCount() {
			break
		}
	}

	return c
}

func ResolveSecretShortID(id string) string {
	if len(id) == 8 {
		// TODO do a real cache
		cache := buildSecretShortIDCache()
		nlid, ok := cache[id]
		if ok {
			if len(nlid) == 1 {
				return nlid[0]
			} else {
				return "local-short-id-conflict"
			}
		}
		return id
	} else {
		return id
	}
}
