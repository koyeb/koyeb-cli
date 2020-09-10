package koyeb

import (
	"fmt"

	"github.com/go-openapi/swag"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	secrets "github.com/koyeb/koyeb-cli/pkg/kclient/client/secrets"
	apimodel "github.com/koyeb/koyeb-cli/pkg/kclient/models"
)

var (
	createSecretCommand = &cobra.Command{
		Use:     "secrets [resource]",
		Aliases: []string{"secret"},
		Short:   "Create secrets",
		RunE:    createSecrets,
	}
	getSecretCommand = &cobra.Command{
		Use:     "secrets [resource]",
		Aliases: []string{"secret"},
		Short:   "Get secrets",
		RunE:    getSecrets,
	}
	describeSecretCommand = &cobra.Command{
		Use:     "secrets [resource]",
		Aliases: []string{"secret"},
		Short:   "Describe secrets",
		RunE:    getSecrets,
	}
	updateSecretCommand = &cobra.Command{
		Use:     "secrets [resource]",
		Aliases: []string{"secret"},
		Short:   "Update secrets",
		RunE:    updateSecrets,
	}
	deleteSecretCommand = &cobra.Command{
		Use:     "secrets [resource]",
		Aliases: []string{"secret"},
		Short:   "Delete secrets",
		RunE:    deleteSecrets,
	}
)

type StorageSecretsBody struct {
	Secrets []StorageSecret `json:"secrets"`
}

func (a *StorageSecretsBody) MarshalBinary() ([]byte, error) {
	if len(a.Secrets) == 1 {
		return swag.WriteJSON(a.Secrets[0])
	} else {
		return swag.WriteJSON(a)
	}
}

func (a *StorageSecretsBody) GetHeaders() []string {
	return []string{"id", "name", "value", "updated_at"}
}

func (a *StorageSecretsBody) GetTableFields() [][]string {
	var data [][]string
	for _, item := range a.Secrets {
		var fields []string
		for _, field := range a.GetHeaders() {
			fields = append(fields, item.GetField(field))
		}
		data = append(data, fields)
	}
	return data
}

func (a *StorageSecretsBody) New() interface{} {
	return &apimodel.StorageSecret{}
}

func (a *StorageSecretsBody) Append(item interface{}) {
	secret := item.(*apimodel.StorageSecret)
	a.Secrets = append(a.Secrets, StorageSecret{*secret})
}

type StorageSecret struct {
	apimodel.StorageSecret
}

func (a StorageSecret) GetNewBody() *apimodel.StorageNewSecret {
	newBody := apimodel.StorageNewSecret{}
	copier.Copy(&newBody, &a.StorageSecret)
	return &newBody
}

func (a StorageSecret) GetUpdateBody() *apimodel.StorageSecret {
	updateBody := apimodel.StorageSecret{}
	copier.Copy(&updateBody, &a.StorageSecret)
	return &updateBody
}

func (a StorageSecret) GetField(field string) string {

	type StorageSecret struct {
		apimodel.StorageSecret
	}

	return getField(a.StorageSecret, field)
}

func displaySecrets(items []*apimodel.StorageSecret, format string) {
	var secrets StorageSecretsBody

	for _, item := range items {
		secrets.Secrets = append(secrets.Secrets, StorageSecret{*item})
	}
	render(&secrets, format)
}

func getSecrets(cmd *cobra.Command, args []string) error {
	client := getApiClient()

	var all []*apimodel.StorageSecret

	if len(args) > 0 {
		for _, arg := range args {
			p := secrets.NewSecretsGetSecretParams()
			p.ID = arg
			resp, err := client.Secrets.SecretsGetSecret(p, getAuth())
			if err != nil {
				apiError(err)
				continue
			}
			log.Debugf("got response: %v", resp)
			all = append([]*apimodel.StorageSecret{resp.GetPayload().Secret}, all...)
		}

	} else {
		page := 0
		limit := 10
		offset := 0

		for {
			p := secrets.NewSecretsListSecretsParams()
			strLimit := fmt.Sprintf("%d", limit)
			p.SetLimit(&strLimit)
			strOffset := fmt.Sprintf("%d", offset)
			p.SetOffset(&strOffset)

			resp, err := client.Secrets.SecretsListSecrets(p, getAuth())
			if err != nil {
				apiError(err)
				er(err)
			}
			log.Debugf("got response: %v", resp)
			all = append(resp.GetPayload().Secrets, all...)
			page += 1
			offset = page * limit
			if int64(offset) >= resp.GetPayload().Count {
				break
			}
		}

	}
	format := "table"
	if cmd.Parent().Name() == "describe" {
		format = "yaml"
	}
	displaySecrets(all, format)

	return nil
}

func createSecrets(cmd *cobra.Command, args []string) error {
	var all StorageSecretsBody

	log.Debugf("Loading file %s", file)
	err := loadMultiple(file, &all, "secrets")
	if err != nil {
		er(err)
	}
	log.Debugf("Content loaded %v", all.Secrets)

	client := getApiClient()
	for _, secret := range all.Secrets {
		p := secrets.NewSecretsNewSecretParams()
		p.SetBody(secret.GetNewBody())
		resp, err := client.Secrets.SecretsNewSecret(p, getAuth())
		if err != nil {
			apiError(err)
			continue
		}
		log.Debugf("got response: %v", resp)
	}
	return nil
}

func updateSecrets(cmd *cobra.Command, args []string) error {
	var all StorageSecretsBody

	log.Debugf("Loading file %s", file)
	err := loadMultiple(file, &all, "secrets")
	if err != nil {
		er(err)
	}
	log.Debugf("Content loaded %v", all.Secrets)

	client := getApiClient()
	for _, st := range all.Secrets {
		p := secrets.NewSecretsUpdateSecretParams()
		updateBody := st.GetUpdateBody()
		if updateBody.ID != "" {
			p.SetID(updateBody.ID)
		} else {
			p.SetID(updateBody.Name)
		}
		p.SetBody(st.GetUpdateBody())
		resp, err := client.Secrets.SecretsUpdateSecret(p, getAuth())
		if err != nil {
			apiError(err)
			continue
		}
		log.Debugf("got response: %v", resp)
	}
	return nil
}

func deleteSecrets(cmd *cobra.Command, args []string) error {
	client := getApiClient()

	if len(args) > 0 {
		for _, arg := range args {
			p := secrets.NewSecretsDeleteSecretParams()
			p.ID = arg
			resp, err := client.Secrets.SecretsDeleteSecret(p, getAuth())
			if err != nil {
				apiError(err)
				continue
			}
			log.Debugf("got response: %v", resp)
			log.Infof("Secret %s deleted", p.ID)
		}
	}
	return nil
}
