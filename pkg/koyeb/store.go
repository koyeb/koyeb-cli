package koyeb

import (
	"fmt"

	"github.com/go-openapi/swag"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/koyeb/koyeb-cli/pkg/gen/kclient/client/store"
	apimodel "github.com/koyeb/koyeb-cli/pkg/gen/kclient/models"
)

var (
	createStoreCommand = &cobra.Command{
		Use:     "stores [resource]",
		Aliases: []string{"store"},
		Short:   "Create stores",
		RunE:    createStores,
	}
	getStoreCommand = &cobra.Command{
		Use:     "stores [resource]",
		Aliases: []string{"store"},
		Short:   "Get stores",
		RunE:    getStores,
	}
	describeStoreCommand = &cobra.Command{
		Use:     "stores [resource]",
		Aliases: []string{"store"},
		Short:   "Describe stores",
		RunE:    getStores,
	}
	updateStoreCommand = &cobra.Command{
		Use:     "stores [resource]",
		Aliases: []string{"store"},
		Short:   "Update stores",
		RunE:    updateStores,
	}
	deleteStoreCommand = &cobra.Command{
		Use:     "stores [resource]",
		Aliases: []string{"store"},
		Short:   "Delete stores",
		RunE:    deleteStores,
	}
)

type StorageStoresBody struct {
	Stores []StorageStore `json:"stores"`
}

func (a *StorageStoresBody) MarshalBinary() ([]byte, error) {
	if len(a.Stores) == 1 {
		return swag.WriteJSON(a.Stores[0])
	} else {
		return swag.WriteJSON(a)
	}
}

func (a *StorageStoresBody) GetTable() TableInfo {
	res := TableInfo{
		headers: []string{"id", "name", "type", "region", "status", "updated_at"},
	}
	for _, item := range a.Stores {
		var fields []string
		for _, field := range res.headers {
			fields = append(fields, item.GetField(field))
		}
		res.fields = append(res.fields, fields)
	}
	return res
}

func (a *StorageStoresBody) New() interface{} {
	return &apimodel.StorageStore{}
}

func (a *StorageStoresBody) Append(item interface{}) {
	store := item.(*apimodel.StorageStore)
	a.Stores = append(a.Stores, StorageStore{*store})
}

type StorageStore struct {
	apimodel.StorageStore
}

func (a StorageStore) GetNewBody() *apimodel.StorageNewStore {
	newBody := apimodel.StorageNewStore{}
	copier.Copy(&newBody, &a.StorageStore)
	return &newBody
}

func (a StorageStore) GetUpdateBody() *apimodel.StorageStore {
	updateBody := apimodel.StorageStore{}
	copier.Copy(&updateBody, &a.StorageStore)
	return &updateBody
}

func (a StorageStore) GetField(field string) string {

	type StorageStore struct {
		apimodel.StorageStore
	}

	return getField(a.StorageStore, field)
}

func displayStores(items []*apimodel.StorageStore, format string) {
	var stores StorageStoresBody

	for _, item := range items {
		stores.Stores = append(stores.Stores, StorageStore{*item})
	}
	render(&stores, format)
}

func getStores(cmd *cobra.Command, args []string) error {
	client := getApiClient()

	var all []*apimodel.StorageStore

	if len(args) > 0 {
		for _, arg := range args {
			p := store.NewStoreGetStoreParams()
			p.ID = arg
			resp, err := client.Store.StoreGetStore(p, getAuth())
			if err != nil {
				logApiError(err)
				continue
			}
			log.Debugf("got response: %v", resp)
			all = append([]*apimodel.StorageStore{resp.GetPayload().Store}, all...)
		}

	} else {
		page := 0
		limit := 10
		offset := 0

		for {
			p := store.NewStoreListStoresParams()
			strLimit := fmt.Sprintf("%d", limit)
			p.SetLimit(&strLimit)
			strOffset := fmt.Sprintf("%d", offset)
			p.SetOffset(&strOffset)

			resp, err := client.Store.StoreListStores(p, getAuth())
			if err != nil {
				fatalApiError(err)
			}
			log.Debugf("got response: %v", resp)
			all = append(resp.GetPayload().Stores, all...)
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
	displayStores(all, format)

	return nil
}

func createStores(cmd *cobra.Command, args []string) error {
	var all StorageStoresBody

	log.Debugf("Loading file %s", file)
	err := loadMultiple(file, &all, "stores")
	if err != nil {
		er(err)
	}
	log.Debugf("Content loaded %v", all.Stores)

	client := getApiClient()
	for _, st := range all.Stores {
		p := store.NewStoreNewStoreParams()
		p.SetBody(st.GetNewBody())
		resp, err := client.Store.StoreNewStore(p, getAuth())
		if err != nil {
			logApiError(err)
			continue
		}
		log.Debugf("got response: %v", resp)
	}
	return nil
}

func updateStores(cmd *cobra.Command, args []string) error {
	var all StorageStoresBody

	log.Debugf("Loading file %s", file)
	err := loadMultiple(file, &all, "stores")
	if err != nil {
		er(err)
	}
	log.Debugf("Content loaded %v", all.Stores)

	client := getApiClient()
	for _, st := range all.Stores {
		p := store.NewStoreUpdateStoreParams()
		updateBody := st.GetUpdateBody()
		if updateBody.ID != "" {
			p.SetID(updateBody.ID)
		} else {
			p.SetID(updateBody.Name)
		}
		p.SetBody(st.GetUpdateBody())
		resp, err := client.Store.StoreUpdateStore(p, getAuth())
		if err != nil {
			logApiError(err)
			continue
		}
		log.Debugf("got response: %v", resp)
	}
	return nil
}

func deleteStores(cmd *cobra.Command, args []string) error {
	client := getApiClient()

	if len(args) > 0 {
		for _, arg := range args {
			p := store.NewStoreDeleteStoreParams()
			p.ID = arg
			resp, err := client.Store.StoreDeleteStore(p, getAuth())
			if err != nil {
				logApiError(err)
				continue
			}
			log.Debugf("got response: %v", resp)
			log.Infof("Store %s deleted", p.ID)
		}
	}
	return nil
}
