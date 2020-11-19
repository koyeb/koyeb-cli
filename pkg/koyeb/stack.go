package koyeb

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-openapi/swag"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/koyeb/koyeb-cli/pkg/gen/kclient/client/stack"
	apimodel "github.com/koyeb/koyeb-cli/pkg/gen/kclient/models"
)

var (
	newStackName       string
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
	logsStackEventsCommand = &cobra.Command{
		Use:     "stack-events [stack]",
		Aliases: []string{"stack-event"},
		Short:   "Logs stack events",
		Args:    cobra.MinimumNArgs(1),
		RunE:    logStackEvents,
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

func (a *StorageStacksBody) GetTable() TableInfo {
	res := TableInfo{
		headers: []string{"id", "name", "status", "updated_at"},
	}
	for _, item := range a.Stacks {
		var fields []string
		for _, field := range res.headers {
			fields = append(fields, item.GetField(field))
		}
		res.fields = append(res.fields, fields)
	}
	return res
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

func (a *StorageStacksDetailBody) GetTable() TableInfo {
	res := TableInfo{
		headers: []string{"id", "name", "status", "latest_revision_sha", "deployed_revision_sha", "updated_at"},
	}
	for _, item := range a.Stacks {
		var fields []string
		for _, field := range res.headers {
			fields = append(fields, item.GetField(field))
		}
		res.fields = append(res.fields, fields)
	}
	return res
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
				logApiError(err)
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
				fatalApiError(err)
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
		fmt.Printf("Stack:\n")
		format = "yaml"
	}
	if len(args) > 0 {
		displayStacksDetail(all, format)
		if cmd.Parent().Name() == "describe" {
			fmt.Printf("Functions:\n")
			getStackFunctions(cmd.Parent(), args)
		}
	} else {
		displayStacks(all, format)
	}

	return nil
}

func createStacks(cmd *cobra.Command, args []string) error {
	var all StorageStacksUpsertBody

	if file != "" {
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
				logApiError(err)
				continue
			}
			log.Debugf("got response: %v", resp)
		}
	} else if newStackName != "" {
		client := getApiClient()
		p := stack.NewStackNewStackParams()
		var stack StorageStackUpsert
		st := stack.GetNewBody()
		st.Name = newStackName
		p.SetBody(st)
		resp, err := client.Stack.StackNewStack(p, getAuth())
		if err != nil {
			logApiError(err)
			return nil
		}
		log.Debugf("got response: %v", resp)
	} else {
		return errors.New("Missing file or name")
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
			logApiError(err)
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
				logApiError(err)
				continue
			}
			log.Debugf("got response: %v", resp)
			log.Infof("Stack %s deleted", p.ID)
		}
	}
	return nil
}

type EventMessageResult struct {
	Message   string
	Id        string
	CreatedAt string
	Data      string
	Event     interface{}
}

type EventMessage struct {
	Result EventMessageResult
}

func (l EventMessage) String() string {
	return fmt.Sprintf("%s data:%v event:%v", l.Result.Message, l.Result.Data, l.Result.Event)
}

func logStackEvents(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		// Nothing to do
		return nil
	}

	path := fmt.Sprintf("/v1/stacks/%s/events/tail", args[0])

	u, err := url.Parse(apiurl)
	if err != nil {
		er(err)
	}

	u.Path = path
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}

	h := http.Header{"Sec-Websocket-Protocol": []string{fmt.Sprintf("Bearer, %s", token)}}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), h)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			msg := EventMessage{}
			err := c.ReadJSON(&msg)
			if err != nil {
				log.Println("error:", err)
				return
			}
			log.Debugf("%v", msg.Result)
			log.Printf("%s", msg)
		}
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return nil
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.PingMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return err
			}
		}
	}
}
