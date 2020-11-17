package koyeb

import (
	"fmt"

	"github.com/go-openapi/swag"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/koyeb/koyeb-cli/pkg/gen/kclient/client/stack"
	apimodel "github.com/koyeb/koyeb-cli/pkg/gen/kclient/models"
)

var (
	createStackRevisionCommand = &cobra.Command{
		Use:     "revisions [stack]",
		Aliases: []string{"revision"},
		Short:   "Create stack revisions",
		Args:    cobra.MinimumNArgs(1),
		RunE:    createStackRevisions,
	}
	getStackRevisionCommand = &cobra.Command{
		Use:     "revisions [stack] [sha]",
		Aliases: []string{"revision"},
		Short:   "Get stack revisions",
		Args:    cobra.MinimumNArgs(1),
		RunE:    getStackRevisions,
	}
	describeStackRevisionCommand = &cobra.Command{
		Use:     "revisions [stack] [sha]",
		Aliases: []string{"revision"},
		Short:   "Describe stack revisions",
		Args:    cobra.MinimumNArgs(2),
		RunE:    getStackRevisions,
	}
)

type StorageStackRevisionsUpsertBody struct {
	StackRevisions []StorageNewStackRevisionRequest `json:"stackrevisions"`
}

func (a *StorageStackRevisionsUpsertBody) Append(item interface{}) {
	stack := item.(*apimodel.StorageNewStackRevisionRequest)
	a.StackRevisions = append(a.StackRevisions, StorageNewStackRevisionRequest{*stack})
}

func (a *StorageStackRevisionsUpsertBody) New() interface{} {
	return &apimodel.StorageNewStackRevisionRequest{}
}

type StorageStackRevisionsBody struct {
	StackRevisions []StorageStackRevision `json:"stackrevisions"`
}

func (a *StorageStackRevisionsBody) MarshalBinary() ([]byte, error) {
	if len(a.StackRevisions) == 1 {
		return swag.WriteJSON(a.StackRevisions[0])
	} else {
		return swag.WriteJSON(a)
	}
}

func (a *StorageStackRevisionsBody) GetHeaders() []string {
	return []string{"sha", "commit_info.message", "status", "created_at"}
}

func (a *StorageStackRevisionsBody) GetTableFields() [][]string {
	var data [][]string
	for _, item := range a.StackRevisions {
		var fields []string
		for _, field := range a.GetHeaders() {
			fields = append(fields, item.GetField(field))
		}
		data = append(data, fields)
	}
	return data
}

func (a *StorageStackRevisionsBody) New() interface{} {
	return &apimodel.StorageStackRevision{}
}

func (a *StorageStackRevisionsBody) Append(item interface{}) {
	stack := item.(*apimodel.StorageStackRevision)
	a.StackRevisions = append(a.StackRevisions, StorageStackRevision{*stack})
}

type StorageStackRevisionsDetailBody struct {
	StorageStackRevisionsBody
}

func (a *StorageStackRevisionsDetailBody) GetHeaders() []string {
	return []string{"sha", "commit_info.message", "status", "created_at"}
}

func (a *StorageStackRevisionsDetailBody) GetTableFields() [][]string {
	var data [][]string
	for _, item := range a.StackRevisions {
		var fields []string
		for _, field := range a.GetHeaders() {
			fields = append(fields, item.GetField(field))
		}
		data = append(data, fields)
	}
	return data
}

type StorageStackRevision struct {
	apimodel.StorageStackRevision
}

func (a StorageStackRevision) GetField(field string) string {

	type StorageStackRevision struct {
		apimodel.StorageStackRevision
	}

	return getField(a.StorageStackRevision, field)
}

type StorageNewStackRevisionRequest struct {
	apimodel.StorageNewStackRevisionRequest
}

func (a StorageNewStackRevisionRequest) GetNewBody() *apimodel.StorageNewStackRevisionRequest {
	newBody := apimodel.StorageNewStackRevisionRequest{}
	copier.Copy(&newBody, &a.StorageNewStackRevisionRequest)
	return &newBody
}

func (a StorageNewStackRevisionRequest) GetUpdateBody() *apimodel.StorageNewStackRevisionRequest {
	updateBody := apimodel.StorageNewStackRevisionRequest{}
	copier.Copy(&updateBody, &a.StorageNewStackRevisionRequest)
	return &updateBody
}

func displayStackRevisions(items []*apimodel.StorageStackRevision, format string) {
	var stackrevisions StorageStackRevisionsBody

	for _, item := range items {
		stackrevisions.StackRevisions = append(stackrevisions.StackRevisions, StorageStackRevision{*item})
	}
	render(&stackrevisions, format)
}

func displayStackRevisionsDetail(items []*apimodel.StorageStackRevision, format string) {
	var stackrevisions StorageStackRevisionsDetailBody

	for _, item := range items {
		stackrevisions.StackRevisions = append(stackrevisions.StackRevisions, StorageStackRevision{*item})
	}
	render(&stackrevisions, format)
}

func getStackRevisions(cmd *cobra.Command, args []string) error {
	client := getApiClient()

	var all []*apimodel.StorageStackRevision

	if len(args) > 1 {
		p := stack.NewStackGetStackRevisionParams()
		p.StackID = args[0]
		p.Sha = args[1]
		resp, err := client.Stack.StackGetStackRevision(p, getAuth())
		if err != nil {
			fatalApiError(err)
		}
		log.Debugf("got response: %v", resp)
		st := resp.GetPayload().Revision
		all = append([]*apimodel.StorageStackRevision{st}, all...)
	} else {
		page := 0
		limit := 10
		offset := 0

		for {
			p := stack.NewStackListStackRevisionsParams()
			strLimit := fmt.Sprintf("%d", limit)
			p.SetLimit(&strLimit)
			strOffset := fmt.Sprintf("%d", offset)
			p.SetOffset(&strOffset)
			p.SetStackID(args[0])

			resp, err := client.Stack.StackListStackRevisions(p, getAuth())
			if err != nil {
				fatalApiError(err)
			}
			log.Debugf("got response: %v", resp)
			var stackrevisions []*apimodel.StorageStackRevision

			for _, st := range resp.GetPayload().Revisions {
				stack := &apimodel.StorageStackRevision{
					Sha:        st.Sha,
					CommitInfo: st.CommitInfo,
					Status:     st.Status,
					CreatedAt:  st.CreatedAt,
				}
				stackrevisions = append(stackrevisions, stack)
			}
			all = append(stackrevisions, all...)
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
	if len(args) > 1 {
		displayStackRevisionsDetail(all, format)
	} else {
		displayStackRevisions(all, format)
	}

	return nil
}

func createStackRevisions(cmd *cobra.Command, args []string) error {
	var all StorageStackRevisionsUpsertBody

	log.Debugf("Loading file %s", file)
	yml, err := loadYaml(file)
	if err != nil {
		er(err)
	}
	log.Debugf("Content loaded %v", all.StackRevisions)

	client := getApiClient()
	p := stack.NewStackNewStackRevisionParams()
	p.SetStackID(args[0])
	var r StorageNewStackRevisionRequest
	rev := r.GetNewBody()
	rev.Yaml = yml
	if stackRevisionMessage != "" {
		rev.Message = stackRevisionMessage
	}
	p.SetBody(rev)
	resp, err := client.Stack.StackNewStackRevision(p, getAuth())
	if err != nil {
		fatalApiError(err)
	}
	log.Debugf("got response: %v", resp)

	return nil
}
