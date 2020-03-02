package koyeb

import (
	"fmt"

	"github.com/go-openapi/swag"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	managedstores "github.com/koyeb/koyeb-cli/pkg/kclient/client/managed_stores"
	apimodel "github.com/koyeb/koyeb-cli/pkg/kclient/models"
)

var (
	createManagedStoreCommand = &cobra.Command{
		Use:     "managedstores [resource]",
		Aliases: []string{"ms", "managedstore"},
		Short:   "Create managedstores",
		RunE:    createManagedStores,
	}
	getManagedStoreCommand = &cobra.Command{
		Use:     "managedstores [resource]",
		Aliases: []string{"ms", "managedstore"},
		Short:   "Get managedstores",
		RunE:    getManagedStores,
	}
	describeManagedStoreCommand = &cobra.Command{
		Use:     "managedstores [resource]",
		Aliases: []string{"ms", "managedstore"},
		Short:   "Describe managedstores",
		RunE:    getManagedStores,
	}
	updateManagedStoreCommand = &cobra.Command{
		Use:     "managedstores [resource]",
		Aliases: []string{"ms", "managedstore"},
		Short:   "Update managedstores",
		RunE:    notImplemented,
	}
	deleteManagedStoreCommand = &cobra.Command{
		Use:     "managedstores [resource]",
		Aliases: []string{"ms", "managedstore"},
		Short:   "Delete managedstores",
		RunE:    deleteManagedStores,
	}
)

type StorageManagedStoresBody struct {
	ManagedStores []StorageManagedStoreBody `json:"managedstores"`
}

func (a *StorageManagedStoresBody) MarshalBinary() ([]byte, error) {
	if len(a.ManagedStores) == 1 {
		return swag.WriteJSON(a.ManagedStores[0])
	} else {
		return swag.WriteJSON(a)
	}
}

func (a *StorageManagedStoresBody) GetHeaders() []string {
	return []string{"id", "name", "region", "status", "updated_at"}
}

func (a *StorageManagedStoresBody) GetTableFields() [][]string {
	var data [][]string
	for _, item := range a.ManagedStores {
		var fields []string
		for _, field := range a.GetHeaders() {
			fields = append(fields, item.GetField(field))
		}
		data = append(data, fields)
	}
	return data
}

func (a *StorageManagedStoresBody) New() interface{} {
	return &apimodel.StorageManagedStoreBody{}
}

func (a *StorageManagedStoresBody) Append(item interface{}) {
	managedstore := item.(*apimodel.StorageManagedStoreBody)
	a.ManagedStores = append(a.ManagedStores, StorageManagedStoreBody{*managedstore})
}

type StorageManagedStoreBody struct {
	apimodel.StorageManagedStoreBody
}

func (a StorageManagedStoreBody) GetNewBody() *apimodel.StorageNewManagedStoreBody {
	newBody := apimodel.StorageNewManagedStoreBody{}
	copier.Copy(&newBody, &a.StorageManagedStoreBody)
	return &newBody
}
func (a StorageManagedStoreBody) GetField(field string) string {

	type StorageManagedStoreBody struct {
		apimodel.StorageManagedStoreBody
	}

	return getField(a.StorageManagedStoreBody, field)
}

func displayManagedStores(items []*apimodel.StorageManagedStoreBody, format string) {
	var managedstores StorageManagedStoresBody

	for _, item := range items {
		managedstores.ManagedStores = append(managedstores.ManagedStores, StorageManagedStoreBody{*item})
	}
	render(&managedstores, format)
}

func getManagedStores(cmd *cobra.Command, args []string) error {
	client := getApiClient()

	var all []*apimodel.StorageManagedStoreBody

	if len(args) > 0 {
		for _, arg := range args {
			p := managedstores.NewGetManagedStoreParams()
			p.ID = arg
			resp, err := client.ManagedStores.GetManagedStore(p, getAuth())
			if err != nil {
				apiError(err)
				continue
			}
			log.Debugf("got response: %v", resp)
			all = append([]*apimodel.StorageManagedStoreBody{resp.GetPayload().ManagedStore}, all...)
		}

	} else {
		page := 0
		limit := 10
		offset := 0

		for {
			p := managedstores.NewListManagedStoresParams()
			strLimit := fmt.Sprintf("%d", limit)
			p.SetLimit(&strLimit)
			strOffset := fmt.Sprintf("%d", offset)
			p.SetOffset(&strOffset)

			resp, err := client.ManagedStores.ListManagedStores(p, getAuth())
			if err != nil {
				apiError(err)
				er(err)
			}
			log.Debugf("got response: %v", resp)
			all = append(resp.GetPayload().ManagedStores, all...)
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
	displayManagedStores(all, format)

	return nil
}

func createManagedStores(cmd *cobra.Command, args []string) error {
	var all StorageManagedStoresBody

	log.Debugf("Loading file %s", file)
	err := loadMultiple(file, &all, "managed_stores")
	if err != nil {
		er(err)
	}
	log.Debugf("Content loaded %v", all.ManagedStores)

	client := getApiClient()
	for _, managedstore := range all.ManagedStores {
		p := managedstores.NewNewManagedStoreParams()
		p.SetBody(managedstore.GetNewBody())
		resp, err := client.ManagedStores.NewManagedStore(p, getAuth())
		if err != nil {
			apiError(err)
			continue
		}
		log.Debugf("got response: %v", resp)
	}
	return nil
}

func deleteManagedStores(cmd *cobra.Command, args []string) error {
	client := getApiClient()

	if len(args) > 0 {
		for _, arg := range args {
			p := managedstores.NewDeleteManagedStoreParams()
			p.ID = arg
			resp, err := client.ManagedStores.DeleteManagedStore(p, getAuth())
			if err != nil {
				apiError(err)
				continue
			}
			log.Debugf("got response: %v", resp)
			log.Infof("ManagedStore %s deleted", p.ID)
		}
	}
	return nil
}
