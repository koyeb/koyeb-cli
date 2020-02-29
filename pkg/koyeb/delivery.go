package koyeb

import (
	"fmt"

	"github.com/go-openapi/swag"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/koyeb/koyeb-cli/pkg/kclient/client/deliveries"
	apimodel "github.com/koyeb/koyeb-cli/pkg/kclient/models"
)

var (
	createDeliveryCommand = &cobra.Command{
		Use:     "deliveries [resource]",
		Aliases: []string{"d", "delivery"},
		Short:   "Create deliveries",
		RunE:    createDeliveries,
	}
	getDeliveryCommand = &cobra.Command{
		Use:     "deliveries [resource]",
		Aliases: []string{"d", "delivery"},
		Short:   "Get deliveries",
		RunE:    getDeliveries,
	}
	describeDeliveryCommand = &cobra.Command{
		Use:     "deliveries [resource]",
		Aliases: []string{"d", "delivery"},
		Short:   "Describe deliveries",
		RunE:    getDeliveries,
	}
	updateDeliveryCommand = &cobra.Command{
		Use:     "deliveries [resource]",
		Aliases: []string{"d", "delivery"},
		Short:   "Update deliveries",
		RunE:    notImplemented,
	}
	deleteDeliveryCommand = &cobra.Command{
		Use:     "deliveries [resource]",
		Aliases: []string{"d", "delivery"},
		Short:   "Delete deliveries",
		RunE:    deleteDeliveries,
	}
)

type StorageDeliveriesBody struct {
	Deliveries []StorageDeliveryBody `json:"deliveries"`
}

func (a *StorageDeliveriesBody) MarshalBinary() ([]byte, error) {
	if len(a.Deliveries) == 1 {
		return swag.WriteJSON(a.Deliveries[0])
	} else {
		return swag.WriteJSON(a)
	}
}

func (a *StorageDeliveriesBody) GetHeaders() []string {
	return []string{"id", "name", "endpoint", "status", "updated_at"}
}

func (a *StorageDeliveriesBody) GetTableFields() [][]string {
	var data [][]string
	for _, item := range a.Deliveries {
		var fields []string
		for _, field := range a.GetHeaders() {
			fields = append(fields, item.GetField(field))
		}
		data = append(data, fields)
	}
	return data
}

func (a *StorageDeliveriesBody) New() interface{} {
	return &apimodel.StorageDeliveryBody{}
}

func (a *StorageDeliveriesBody) Append(item interface{}) {
	delivery := item.(*apimodel.StorageDeliveryBody)
	a.Deliveries = append(a.Deliveries, StorageDeliveryBody{*delivery})
}

type StorageDeliveryBody struct {
	apimodel.StorageDeliveryBody
}

func (a StorageDeliveryBody) GetNewBody() *apimodel.StorageNewDeliveryBody {
	newBody := apimodel.StorageNewDeliveryBody{}
	copier.Copy(&newBody, &a.StorageDeliveryBody)
	return &newBody
}
func (a StorageDeliveryBody) GetField(field string) string {

	type StorageDeliveryBody struct {
		apimodel.StorageDeliveryBody
	}

	return getField(a.StorageDeliveryBody, field)
}

func displayDeliveries(items []*apimodel.StorageDeliveryBody, format string) {
	var deliveries StorageDeliveriesBody

	for _, item := range items {
		deliveries.Deliveries = append(deliveries.Deliveries, StorageDeliveryBody{*item})
	}
	render(&deliveries, format)
}

func getDeliveries(cmd *cobra.Command, args []string) error {
	client := getApiClient()

	var all []*apimodel.StorageDeliveryBody

	if len(args) > 0 {
		for _, arg := range args {
			p := deliveries.NewGetDeliveryParams()
			p.ID = arg
			resp, err := client.Deliveries.GetDelivery(p)
			if err != nil {
				apiError(err)
				continue
			}
			log.Debugf("got response: %v", resp)
			all = append([]*apimodel.StorageDeliveryBody{resp.GetPayload().Delivery}, all...)
		}

	} else {
		page := 0
		limit := 10
		offset := 0

		for {
			p := deliveries.NewListDeliveriesParams()
			strLimit := fmt.Sprintf("%d", limit)
			p.SetLimit(&strLimit)
			strOffset := fmt.Sprintf("%d", offset)
			p.SetOffset(&strOffset)

			resp, err := client.Deliveries.ListDeliveries(p)
			if err != nil {
				apiError(err)
				er(err)
			}
			log.Debugf("got response: %v", resp)
			all = append(resp.GetPayload().Deliveries, all...)
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
	displayDeliveries(all, format)

	return nil
}

func createDeliveries(cmd *cobra.Command, args []string) error {
	var all StorageDeliveriesBody

	log.Debugf("Loading file %s", file)
	err := loadMultiple(file, &all, "deliveries")
	if err != nil {
		er(err)
	}
	log.Debugf("Content loaded %v", all.Deliveries)

	client := getApiClient()
	for _, delivery := range all.Deliveries {
		p := deliveries.NewNewDeliveryParams()
		p.SetBody(delivery.GetNewBody())
		resp, err := client.Deliveries.NewDelivery(p)
		if err != nil {
			apiError(err)
			continue
		}
		log.Debugf("got response: %v", resp)
	}
	return nil
}

func deleteDeliveries(cmd *cobra.Command, args []string) error {
	client := getApiClient()

	if len(args) > 0 {
		for _, arg := range args {
			p := deliveries.NewDeleteDeliveryParams()
			p.ID = arg
			resp, err := client.Deliveries.DeleteDelivery(p)
			if err != nil {
				apiError(err)
				continue
			}
			log.Debugf("got response: %v", resp)
			log.Infof("Delivery %s deleted", p.ID)
		}
	}
	return nil
}
