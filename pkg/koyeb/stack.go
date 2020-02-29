package koyeb

import (
	"fmt"

	"github.com/go-openapi/swag"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/koyeb/koyeb-cli/pkg/kclient/client/stacks"
	apimodel "github.com/koyeb/koyeb-cli/pkg/kclient/models"
)

var (
	createStackCommand = &cobra.Command{
		Use:     "stacks [resource]",
		Aliases: []string{"s", "stack"},
		Short:   "Create stacks",
		RunE:    createStacks,
	}
	getStackCommand = &cobra.Command{
		Use:     "stacks [resource]",
		Aliases: []string{"s", "stack"},
		Short:   "Get stacks",
		RunE:    getStacks,
	}
	describeStackCommand = &cobra.Command{
		Use:     "stacks [resource]",
		Aliases: []string{"s", "stack"},
		Short:   "Describe stacks",
		RunE:    getStacks,
	}
	updateStackCommand = &cobra.Command{
		Use:     "stacks [resource]",
		Aliases: []string{"s", "stack"},
		Short:   "Update stacks",
		RunE:    notImplemented,
	}
	deleteStackCommand = &cobra.Command{
		Use:     "stacks [resource]",
		Aliases: []string{"s", "stack"},
		Short:   "Delete stacks",
		RunE:    deleteStacks,
	}
)

type StorageStacksBody struct {
	Stacks []StorageStackBody `json:"stacks"`
}

func (a *StorageStacksBody) MarshalBinary() ([]byte, error) {
	if len(a.Stacks) == 1 {
		return swag.WriteJSON(a.Stacks[0])
	} else {
		return swag.WriteJSON(a)
	}
}

func (a *StorageStacksBody) GetHeaders() []string {
	return []string{"id", "name", "region", "status", "updated_at"}
}

func (a *StorageStacksBody) GetTableFields() [][]string {
	var data [][]string
	for _, item := range a.Stacks {
		var fields []string
		for _, field := range a.GetHeaders() {
			fields = append(fields, item.GetField(field))
		}
		data = append(data, fields)
	}
	return data
}

func (a *StorageStacksBody) New() interface{} {
	return &apimodel.StorageStackBody{}
}

func (a *StorageStacksBody) Append(item interface{}) {
	stack := item.(*apimodel.StorageStackBody)
	a.Stacks = append(a.Stacks, StorageStackBody{*stack})
}

type StorageStackBody struct {
	apimodel.StorageStackBody
}

func (a StorageStackBody) GetNewBody() *apimodel.StorageNewStackBody {
	newBody := apimodel.StorageNewStackBody{}
	copier.Copy(&newBody, &a.StorageStackBody)
	return &newBody
}
func (a StorageStackBody) GetField(field string) string {

	type StorageStackBody struct {
		apimodel.StorageStackBody
	}

	return getField(a.StorageStackBody, field)
}

func displayStacks(items []*apimodel.StorageStackBody, format string) {
	var stacks StorageStacksBody

	for _, item := range items {
		stacks.Stacks = append(stacks.Stacks, StorageStackBody{*item})
	}
	render(&stacks, format)
}

func getStacks(cmd *cobra.Command, args []string) error {
	client := getApiClient()

	var all []*apimodel.StorageStackBody

	if len(args) > 0 {
		for _, arg := range args {
			p := stacks.NewGetStackParams()
			p.ID = arg
			resp, err := client.Stacks.GetStack(p)
			if err != nil {
				apiError(err)
				continue
			}
			log.Debugf("got response: %v", resp)
			all = append([]*apimodel.StorageStackBody{resp.GetPayload().Stack}, all...)
		}

	} else {
		page := 0
		limit := 10
		offset := 0

		for {
			p := stacks.NewListStacksParams()
			strLimit := fmt.Sprintf("%d", limit)
			p.SetLimit(&strLimit)
			strOffset := fmt.Sprintf("%d", offset)
			p.SetOffset(&strOffset)

			resp, err := client.Stacks.ListStacks(p)
			if err != nil {
				apiError(err)
				er(err)
			}
			log.Debugf("got response: %v", resp)
			all = append(resp.GetPayload().Stacks, all...)
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
	displayStacks(all, format)

	return nil
}

func createStacks(cmd *cobra.Command, args []string) error {
	var all StorageStacksBody

	log.Debugf("Loading file %s", file)
	err := loadMultiple(file, &all, "stacks")
	if err != nil {
		er(err)
	}
	log.Debugf("Content loaded %v", all.Stacks)

	client := getApiClient()
	for _, stack := range all.Stacks {
		p := stacks.NewNewStackParams()
		p.SetBody(stack.GetNewBody())
		resp, err := client.Stacks.NewStack(p)
		if err != nil {
			apiError(err)
			continue
		}
		log.Debugf("got response: %v", resp)
	}
	return nil
}

func deleteStacks(cmd *cobra.Command, args []string) error {
	client := getApiClient()

	if len(args) > 0 {
		for _, arg := range args {
			p := stacks.NewDeleteStackParams()
			p.ID = arg
			resp, err := client.Stacks.DeleteStack(p)
			if err != nil {
				apiError(err)
				continue
			}
			log.Debugf("got response: %v", resp)
			log.Infof("Stack %s deleted", p.ID)
		}
	}
	return nil
}
