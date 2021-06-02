package koyeb

import (
	"context"
	"fmt"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

var (
// createSecretCommand = &cobra.Command{
//   Use:     "secrets [resource]",
//   Aliases: []string{"secret"},
//   Short:   "Create secrets",
//   RunE:    createSecrets,
// }
// getSecretCommand = &cobra.Command{
//   Use:     "secrets [resource]",
//   Aliases: []string{"secret"},
//   Short:   "Get secrets",
//   RunE:    getSecrets,
// }
// describeSecretCommand = &cobra.Command{
//   Use:     "secrets [resource]",
//   Aliases: []string{"secret"},
//   Short:   "Describe secrets",
//   RunE:    getSecrets,
// }
// updateSecretCommand = &cobra.Command{
//   Use:     "secrets [resource]",
//   Aliases: []string{"secret"},
//   Short:   "Update secrets",
//   RunE:    updateSecrets,
// }
// deleteSecretCommand = &cobra.Command{
//   Use:     "secrets [resource]",
//   Aliases: []string{"secret"},
//   Short:   "Delete secrets",
//   RunE:    deleteSecrets,
// }
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
		// RunE:    createSecrets,
	}
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
	return secretCmd
}

func NewSecretHandler() *SecretHandler {
	return &SecretHandler{}
}

type SecretHandler struct {
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
			logApiError(err)
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
			logApiError(err)
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

// func (a *StorageSecretsBody) New() interface{} {
//   return &apimodel.StorageSecret{}
// }

// func (a *StorageSecretsBody) Append(item interface{}) {
//   secret := item.(*apimodel.StorageSecret)
//   a.Secrets = append(a.Secrets, StorageSecret{*secret})
// }

// type StorageSecret struct {
//   apimodel.StorageSecret
// }

// func (a StorageSecret) GetNewBody() *apimodel.StorageNewSecret {
//   newBody := apimodel.StorageNewSecret{}
//   copier.Copy(&newBody, &a.StorageSecret)
//   return &newBody
// }

// func (a StorageSecret) GetUpdateBody() *apimodel.StorageSecret {
//   updateBody := apimodel.StorageSecret{}
//   copier.Copy(&updateBody, &a.StorageSecret)
//   return &updateBody
// }

// func (a StorageSecret) GetField(field string) string {

//   type StorageSecret struct {
//     apimodel.StorageSecret
//   }

//   return getField(a.StorageSecret, field)
// }

// func displaySecrets(items *[]koyeb.Secret, format string) {
// secrets := &SecretsTable{items}

// for _, item := range items {
//   secrets.Secrets = append(secrets.Secrets, StorageSecret{*item})
// }
// render(secrets, format)
// }

// func createSecrets(cmd *cobra.Command, args []string) error {
//   var all StorageSecretsBody

//   log.Debugf("Loading file %s", file)
//   err := loadMultiple(file, &all, "secrets")
//   if err != nil {
//     er(err)
//   }
//   log.Debugf("Content loaded %v", all.Secrets)

//   client := getApiClient()
//   for _, secret := range all.Secrets {
//     p := secrets.NewSecretsNewSecretParams()
//     p.SetBody(secret.GetNewBody())
//     resp, err := client.Secrets.SecretsNewSecret(p, getAuth())
//     if err != nil {
//       logApiError(err)
//       continue
//     }
//     log.Debugf("got response: %v", resp)
//   }
//   return nil
// }

// func updateSecrets(cmd *cobra.Command, args []string) error {
//   var all StorageSecretsBody

//   log.Debugf("Loading file %s", file)
//   err := loadMultiple(file, &all, "secrets")
//   if err != nil {
//     er(err)
//   }
//   log.Debugf("Content loaded %v", all.Secrets)

//   client := getApiClient()
//   for _, st := range all.Secrets {
//     p := secrets.NewSecretsUpdateSecretParams()
//     updateBody := st.GetUpdateBody()
//     if updateBody.ID != "" {
//       p.SetID(updateBody.ID)
//     } else {
//       p.SetID(updateBody.Name)
//     }
//     p.SetBody(st.GetUpdateBody())
//     resp, err := client.Secrets.SecretsUpdateSecret(p, getAuth())
//     if err != nil {
//       logApiError(err)
//       continue
//     }
//     log.Debugf("got response: %v", resp)
//   }
//   return nil
// }

// func deleteSecrets(cmd *cobra.Command, args []string) error {
//   client := getApiClient()

//   if len(args) > 0 {
//     for _, arg := range args {
//       p := secrets.NewSecretsDeleteSecretParams()
//       p.ID = arg
//       resp, err := client.Secrets.SecretsDeleteSecret(p, getAuth())
//       if err != nil {
//         logApiError(err)
//         continue
//       }
//       log.Debugf("got response: %v", resp)
//       log.Infof("Secret %s deleted", p.ID)
//     }
//   }
//   return nil
// }
