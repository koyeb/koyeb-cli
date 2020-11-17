package koyeb

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-openapi/swag"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/koyeb/koyeb-cli/pkg/gen/kclient/client/functions"
	"github.com/koyeb/koyeb-cli/pkg/gen/kclient/client/stack"
	apimodel "github.com/koyeb/koyeb-cli/pkg/gen/kclient/models"
)

var (
	getStackFunctionCommand = &cobra.Command{
		Use:     "functions [stack] [name]",
		Aliases: []string{"function"},
		Short:   "Get stack functions",
		Args:    cobra.MinimumNArgs(1),
		RunE:    getStackFunctions,
	}
	describeStackFunctionCommand = &cobra.Command{
		Use:     "functions [stack] [name]",
		Aliases: []string{"function"},
		Short:   "Describe stack functions",
		Args:    cobra.MinimumNArgs(2),
		RunE:    getStackFunctions,
	}
	logsStackFunctionCommand = &cobra.Command{
		Use:     "functions [stack] [name]",
		Aliases: []string{"function"},
		Short:   "Logs stack functions",
		Args:    cobra.MinimumNArgs(2),
		RunE:    logStackFunctions,
	}
	invokeStackFunctionCommand = &cobra.Command{
		Use:     "functions [stack] [name]",
		Aliases: []string{"function"},
		Short:   "Invoke stack functions",
		Args:    cobra.MinimumNArgs(2),
		RunE:    invokeStackFunctions,
	}
)

type StorageFunctionRunInfoHistory struct {
	Runs []StorageFunctionRunInfoListItem `json:"runs"`
}

func (a *StorageFunctionRunInfoHistory) MarshalBinary() ([]byte, error) {
	if len(a.Runs) == 1 {
		return swag.WriteJSON(a.Runs[0])
	} else {
		return swag.WriteJSON(a)
	}
}

func (a *StorageFunctionRunInfoHistory) GetTable() TableInfo {
	res := TableInfo{
		headers: []string{"run_id", "event_id", "fn_name", "state", "start", "end"},
	}
	for _, item := range a.Runs {
		var fields []string
		for _, field := range res.headers {
			if field == "start" || field == "end" {
				fields = append(fields, getField(*item.Executions[0], field))
			} else {
				fields = append(fields, item.GetField(field))
			}
		}
		res.fields = append(res.fields, fields)
	}
	return res
}

func (a *StorageFunctionRunInfoHistory) New() interface{} {
	return &apimodel.StorageFunctionRunInfoListItem{}
}

func (a *StorageFunctionRunInfoHistory) Append(item interface{}) {
	stack := item.(*apimodel.StorageFunctionRunInfoListItem)
	a.Runs = append(a.Runs, StorageFunctionRunInfoListItem{*stack})
}

type StorageStackFunctionsBody struct {
	StackFunctions []StorageStackFunction `json:"stackfunctions"`
}

func (a *StorageStackFunctionsBody) MarshalBinary() ([]byte, error) {
	if len(a.StackFunctions) == 1 {
		return swag.WriteJSON(a.StackFunctions[0])
	} else {
		return swag.WriteJSON(a)
	}
}

func (a *StorageStackFunctionsBody) GetTable() TableInfo {
	res := TableInfo{
		headers: []string{"name", "type"},
	}
	for _, item := range a.StackFunctions {
		var fields []string
		for _, field := range res.headers {
			fields = append(fields, item.GetField(field))
		}
		res.fields = append(res.fields, fields)
	}
	return res
}

func (a *StorageStackFunctionsBody) New() interface{} {
	return &apimodel.StorageFunction{}
}

func (a *StorageStackFunctionsBody) Append(item interface{}) {
	stack := item.(*apimodel.StorageFunction)
	a.StackFunctions = append(a.StackFunctions, StorageStackFunction{*stack})
}

type StorageStackFunctionsDetailBody struct {
	StorageStackFunctionsBody
}

type StorageStackFunction struct {
	apimodel.StorageFunction
}

func (a StorageStackFunction) GetField(field string) string {

	return getField(a.StorageFunction, field)
}

type StorageFunctionRunInfoListItem struct {
	apimodel.StorageFunctionRunInfoListItem
}

func (a StorageFunctionRunInfoListItem) GetField(field string) string {

	return getField(a.StorageFunctionRunInfoListItem, field)
}

func displayStackFunctions(items []*apimodel.StorageFunction, format string) {
	var stackfunctions StorageStackFunctionsBody

	for _, item := range items {
		stackfunctions.StackFunctions = append(stackfunctions.StackFunctions, StorageStackFunction{*item})
	}
	render(&stackfunctions, format)
}

func displayStackFunctionsDetail(items []*apimodel.StorageFunction, format string) {
	var stackfunctions StorageStackFunctionsDetailBody

	for _, item := range items {
		stackfunctions.StackFunctions = append(stackfunctions.StackFunctions, StorageStackFunction{*item})
	}
	render(&stackfunctions, format)
}

func displayStackFunctionsHistory(items []*apimodel.StorageFunctionRunInfoListItem, format string) {
	var stackfunctionsHistory StorageFunctionRunInfoHistory

	for _, item := range items {
		stackfunctionsHistory.Runs = append(stackfunctionsHistory.Runs, StorageFunctionRunInfoListItem{*item})
	}
	render(&stackfunctionsHistory, format)
}

type LogMessageResult struct {
	Message string
}

type LogMessage struct {
	Result LogMessageResult
}

func (l LogMessage) String() string {
	return l.Result.Message
}

func logStackFunctions(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		// Nothing to do
		return nil
	}

	path := fmt.Sprintf("/v1/stacks/%s/revisions/%s/functions/%s/logs/tail", args[0], ":latest", args[1])

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
			msg := LogMessage{}
			err := c.ReadJSON(&msg)
			if err != nil {
				log.Println("error:", err)
				return
			}
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

func invokeStackFunctions(cmd *cobra.Command, args []string) error {
	client := getApiClient()

	if len(args) < 2 {
		// Nothing to do
		return nil
	} else {
		p := functions.NewFunctionsInvokeFunctionParams()
		p.WithStackID(args[0]).WithSha(":latest").WithFunction(args[1])
		if file != "" {
			log.Debugf("Loading file %s", file)
			ev := make(map[string]interface{})
			err := parseFile(file, &ev)
			if err != nil {
				er(err)
			}
			p.SetBody(ev)
		} else {
			log.Debugf("Using default values for event")
			p.SetBody(map[string]string{
				"type":   "debug",
				"source": "debug",
			})
		}
		log.Debugf("Launching debug event %v", p.Body)
		resp, err := client.Functions.FunctionsInvokeFunction(p, getAuth())
		if err != nil {
			fatalApiError(err)
		}
		log.Debugf("got response: %v", resp)
		log.Infof("Event sent: %v", resp.Payload.ID)
	}
	return nil
}

func getStackFunctions(cmd *cobra.Command, args []string) error {
	client := getApiClient()

	var all []*apimodel.StorageFunction

	if len(args) == 0 {
		// Nothing to do
		return nil
	} else if len(args) > 1 {
		p := functions.NewFunctionsGetFunctionParams()
		p.WithStackID(args[0]).WithSha(":latest").WithFunction(args[1])
		resp, err := client.Functions.FunctionsGetFunction(p, getAuth())
		if err != nil {
			fatalApiError(err)
		}
		log.Debugf("got response: %v", resp)
		st := resp.GetPayload().Function
		all = append([]*apimodel.StorageFunction{st}, all...)
	} else {
		r := stack.NewStackGetStackParams()
		r.ID = args[0]
		stack, err := client.Stack.StackGetStack(r, getAuth())
		if err != nil {
			fatalApiError(err)
		}
		if stack.Payload.Stack.LatestRevisionSha == "" {
			log.Debugf("No revision")
			return nil
		}
		p := functions.NewFunctionsListFunctionsParams()
		p = p.WithStackID(args[0]).WithSha(":latest")
		resp, err := client.Functions.FunctionsListFunctions(p, getAuth())
		if err != nil {
			fatalApiError(err)
		}
		log.Debugf("got response: %v", resp)
		var stackfunctions []*apimodel.StorageFunction

		for _, st := range resp.GetPayload().Functions {
			stack := &apimodel.StorageFunction{
				Name: st.Name,
				Type: st.Type,
			}
			stackfunctions = append(stackfunctions, stack)
		}
		all = append(stackfunctions, all...)

	}
	format := "table"
	if cmd.Parent().Name() == "describe" {
		fmt.Printf("Function:\n")
		format = "yaml"
	}
	if len(args) > 1 {
		displayStackFunctionsDetail(all, format)

		if cmd.Parent().Name() == "describe" {
			fmt.Printf("History:\n")
			// display history
			err := getStackFunctionHistory(cmd, args)
			if err != nil {
				return err
			}
		}
	} else {
		displayStackFunctions(all, format)
	}

	return nil
}

func getStackFunctionHistory(cmd *cobra.Command, args []string) error {
	client := getApiClient()

	var all []*apimodel.StorageFunctionRunInfoListItem

	if len(args) != 2 {
		// Nothing to do
		return nil
	} else {
		p := functions.NewFunctionsFetchFunctionExecutionsParams()
		p = p.WithStackID(args[0]).WithSha(":latest").WithFunction(args[1])
		resp, err := client.Functions.FunctionsFetchFunctionExecutions(p, getAuth())
		if err != nil {
			fatalApiError(err)
		}
		log.Debugf("got response: %v", resp)
		var stackfunctionsHistory []*apimodel.StorageFunctionRunInfoListItem

		for _, st := range resp.GetPayload().Executions {
			stackfunctionsHistory = append(stackfunctionsHistory, st)
		}
		all = append(stackfunctionsHistory, all...)

	}
	format := "table"
	displayStackFunctionsHistory(all, format)

	return nil
}
