package koyeb

import (
	"fmt"

	"github.com/go-openapi/swag"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	stack "github.com/koyeb/koyeb-cli/pkg/kclient/client/stack"
	apimodel "github.com/koyeb/koyeb-cli/pkg/kclient/models"
)

var (
	createStackCommand = &cobra.Command{
		Use:     "stacks [resource]",
		Aliases: []string{"stack"},
		Short:   "Create stacks",
		RunE:    createStacks,
	}
	getStackCommand = &cobra.Command{
		Use:     "stacks [resource]",
		Aliases: []string{"stack"},
		Short:   "Get stacks",
		RunE:    getStacks,
	}
	describeStackCommand = &cobra.Command{
		Use:     "stacks [resource]",
		Aliases: []string{"stack"},
		Short:   "Describe stacks",
		RunE:    getStacks,
	}
	updateStackCommand = &cobra.Command{
		Use:     "stacks [resource]",
		Aliases: []string{"stack"},
		Short:   "Update stacks",
		RunE:    updateStacks,
	}
	deleteStackCommand = &cobra.Command{
		Use:     "stacks [resource]",
		Aliases: []string{"stack"},
		Short:   "Delete stacks",
		RunE:    deleteStacks,
	}
)

type StorageStacksUpsertBody struct {
	Stacks []StorageStackUpsert `json:"stacks"`
}

func (a *StorageStacksUpsertBody) Append(item interface{}) {
	stack := item.(*apimodel.StorageStackUpsert)
	a.Stacks = append(a.Stacks, StorageStackUpsert{*stack})
}

func (a *StorageStacksUpsertBody) New() interface{} {
	return &apimodel.StorageStackUpsert{}
}

type StorageStacksBody struct {
	Stacks []StorageStack `json:"stacks"`
}

func (a *StorageStacksBody) MarshalBinary() ([]byte, error) {
	if len(a.Stacks) == 1 {
		return swag.WriteJSON(a.Stacks[0])
	} else {
		return swag.WriteJSON(a)
	}
}

func (a *StorageStacksBody) GetHeaders() []string {
	return []string{"id", "name", "status", "updated_at"}
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
	return &apimodel.StorageStack{}
}

func (a *StorageStacksBody) Append(item interface{}) {
	stack := item.(*apimodel.StorageStack)
	a.Stacks = append(a.Stacks, StorageStack{*stack})
}

type StorageStacksDetailBody struct {
	StorageStacksBody
}

func (a *StorageStacksDetailBody) GetHeaders() []string {
	return []string{"id", "name", "status", "latest_revision_sha", "deployed_revision_sha", "updated_at"}
}

func (a *StorageStacksDetailBody) GetTableFields() [][]string {
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

type StorageStack struct {
	apimodel.StorageStack
}

func (a StorageStack) GetField(field string) string {

	type StorageStack struct {
		apimodel.StorageStack
	}

	return getField(a.StorageStack, field)
}

type StorageStackUpsert struct {
	apimodel.StorageStackUpsert
}

func (a StorageStackUpsert) GetNewBody() *apimodel.StorageStackUpsert {
	newBody := apimodel.StorageStackUpsert{}
	copier.Copy(&newBody, &a.StorageStackUpsert)
	return &newBody
}

func (a StorageStackUpsert) GetUpdateBody() *apimodel.StorageStackUpsert {
	updateBody := apimodel.StorageStackUpsert{}
	copier.Copy(&updateBody, &a.StorageStackUpsert)
	return &updateBody
}

func displayStacks(items []*apimodel.StorageStack, format string) {
	var stacks StorageStacksBody

	for _, item := range items {
		stacks.Stacks = append(stacks.Stacks, StorageStack{*item})
	}
	render(&stacks, format)
}

func displayStacksDetail(items []*apimodel.StorageStack, format string) {
	var stacks StorageStacksDetailBody

	for _, item := range items {
		stacks.Stacks = append(stacks.Stacks, StorageStack{*item})
	}
	render(&stacks, format)
}

func getStacks(cmd *cobra.Command, args []string) error {
	client := getApiClient()

	var all []*apimodel.StorageStack

	if len(args) > 0 {
		for _, arg := range args {
			p := stack.NewStackGetStackParams()
			p.ID = arg
			resp, err := client.Stack.StackGetStack(p, getAuth())
			if err != nil {
				apiError(err)
				continue
			}
			log.Debugf("got response: %v", resp)
			st := resp.GetPayload().Stack
			all = append([]*apimodel.StorageStack{st}, all...)
		}

	} else {
		page := 0
		limit := 10
		offset := 0

		for {
			p := stack.NewStackListStacksParams()
			strLimit := fmt.Sprintf("%d", limit)
			p.SetLimit(&strLimit)
			strOffset := fmt.Sprintf("%d", offset)
			p.SetOffset(&strOffset)

			resp, err := client.Stack.StackListStacks(p, getAuth())
			if err != nil {
				apiError(err)
				er(err)
			}
			log.Debugf("got response: %v", resp)
			var stacks []*apimodel.StorageStack

			for _, st := range resp.GetPayload().Stacks {
				stack := &apimodel.StorageStack{
					ID:                  st.ID,
					Name:                st.Name,
					LatestRevisionSha:   st.LatestRevisionSha,
					DeployedRevisionSha: st.DeployedRevisionSha,
					Status:              st.Status,
					Repository:          st.Repository,
					UpdatedAt:           st.UpdatedAt,
					CreatedAt:           st.CreatedAt,
				}
				stacks = append(stacks, stack)
			}
			all = append(stacks, all...)
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
	if len(args) > 0 {
		displayStacksDetail(all, format)
	} else {
		displayStacks(all, format)
	}

	return nil
}

func createStacks(cmd *cobra.Command, args []string) error {
	var all StorageStacksUpsertBody

	log.Debugf("Loading file %s", file)
	err := loadMultiple(file, &all, "stacks")
	if err != nil {
		er(err)
	}
	log.Debugf("Content loaded %v", all.Stacks)

	client := getApiClient()
	for _, st := range all.Stacks {
		p := stack.NewStackNewStackParams()
		p.SetBody(st.GetNewBody())
		resp, err := client.Stack.StackNewStack(p, getAuth())
		if err != nil {
			apiError(err)
			continue
		}
		log.Debugf("got response: %v", resp)
	}
	return nil
}

func updateStacks(cmd *cobra.Command, args []string) error {
	var all StorageStacksUpsertBody

	log.Debugf("Loading file %s", file)
	err := loadMultiple(file, &all, "stacks")
	if err != nil {
		er(err)
	}
	log.Debugf("Content loaded %v", all.Stacks)

	client := getApiClient()
	for _, st := range all.Stacks {
		p := stack.NewStackUpdateStackParams()
		updateBody := st.GetUpdateBody()
		p.SetID(updateBody.Name)
		p.SetBody(st.GetUpdateBody())
		resp, err := client.Stack.StackUpdateStack(p, getAuth())
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
			p := stack.NewStackDeleteStackParams()
			p.ID = arg
			resp, err := client.Stack.StackDeleteStack(p, getAuth())
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
