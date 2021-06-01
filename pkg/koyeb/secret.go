package koyeb

import (
	"context"
	// "fmt"
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
		// RunE:    createSecrets,
	}
	secretCmd.AddCommand(getSecretCmd)

	listSecretCmd := &cobra.Command{
		Use:   "list",
		Short: "List secrets",
		RunE:  listSecrets,
	}
	secretCmd.AddCommand(listSecretCmd)
	describeSecretCmd := &cobra.Command{
		Use:   "describe [name]",
		Short: "Describe secrets",
		RunE:  describeSecrets,
	}
	secretCmd.AddCommand(describeSecretCmd)
	return secretCmd
}

type GetSecretReply struct {
	koyeb.GetSecretReply
}

func (a *GetSecretReply) MarshalBinary() ([]byte, error) {
	return a.GetSecretReply.GetSecret().MarshalJSON()
}

func (a *GetSecretReply) GetTable() TableInfo {
	res := TableInfo{
		headers: []string{"id", "name", "value", "updated_at"},
	}
	item := a.GetSecret()
	var fields []string
	fields = append(fields, item.GetId(), item.GetName(), item.GetValue(), item.GetUpdatedAt().String())
	res.fields = append(res.fields, fields)
	return res
}

type ListSecretsReply struct {
	koyeb.ListSecretsReply
}

func (a *ListSecretsReply) MarshalBinary() ([]byte, error) {
	return a.ListSecretsReply.MarshalJSON()
}

func (a *ListSecretsReply) GetTable() TableInfo {
	res := TableInfo{
		headers: []string{"id", "name", "value", "updated_at"},
	}
	for _, item := range a.GetSecrets() {
		var fields []string
		fields = append(fields, item.GetId(), item.GetName(), item.GetValue(), item.GetUpdatedAt().String())
		res.fields = append(res.fields, fields)
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

func getSecretsFormat(cmd *cobra.Command, args []string, format string) error {
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

func listSecretsFormat(cmd *cobra.Command, args []string, format string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	res, _, err := client.SecretsApi.ListSecrets(ctx).Execute()
	if err != nil {
		logApiError(err)
	}

	render(format, &ListSecretsReply{res})

	return nil
}

func listSecrets(cmd *cobra.Command, args []string) error {
	format := "table"
	return listSecretsFormat(cmd, args, format)
}

func describeSecrets(cmd *cobra.Command, args []string) error {
	format := "yaml"
	if len(args) == 0 {
		return listSecretsFormat(cmd, args, format)
	}
	return getSecretsFormat(cmd, args, format)
}

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
