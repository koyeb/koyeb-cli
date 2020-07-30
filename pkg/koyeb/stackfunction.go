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

	"github.com/koyeb/koyeb-cli/pkg/kclient/client/functions"
	apimodel "github.com/koyeb/koyeb-cli/pkg/kclient/models"
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
		Args:    cobra.MinimumNArgs(1),
		RunE:    getStackFunctions,
	}
	logsStackFunctionCommand = &cobra.Command{
		Use:     "functions [stack] [name]",
		Aliases: []string{"function"},
		Short:   "Logs stack functions",
		Args:    cobra.MinimumNArgs(1),
		RunE:    logStackFunctions,
	}
	runStackFunctionCommand = &cobra.Command{
		Use:     "functions [stack] [name]",
		Aliases: []string{"function"},
		Short:   "Run stack functions",
		Args:    cobra.MinimumNArgs(1),
		RunE:    runStackFunctions,
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

func (a *StorageFunctionRunInfoHistory) GetHeaders() []string {
	return []string{"run_id", "event_id", "fn_name", "state", "start", "end"}
}

func (a *StorageFunctionRunInfoHistory) GetTableFields() [][]string {
	var data [][]string
	for _, item := range a.Runs {
		var fields []string
		for _, field := range a.GetHeaders() {
			fields = append(fields, item.GetField(field))
		}
		data = append(data, fields)
	}
	return data
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

func (a *StorageStackFunctionsBody) GetHeaders() []string {
	return []string{"name", "type"}
}

func (a *StorageStackFunctionsBody) GetTableFields() [][]string {
	var data [][]string
	for _, item := range a.StackFunctions {
		var fields []string
		for _, field := range a.GetHeaders() {
			fields = append(fields, item.GetField(field))
		}
		data = append(data, fields)
	}
	return data
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

func runStackFunctions(cmd *cobra.Command, args []string) error {
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
			apiError(err)
			return err
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
			apiError(err)
			return err
		}
		log.Debugf("got response: %v", resp)
		st := resp.GetPayload().Function
		all = append([]*apimodel.StorageFunction{st}, all...)
	} else {
		p := functions.NewFunctionsListFunctionsParams()
		p = p.WithStackID(args[0]).WithSha(":latest")
		resp, err := client.Functions.FunctionsListFunctions(p, getAuth())
		if err != nil {
			apiError(err)
			return err
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
			getStackFunctionHistory(cmd, args)
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
		p := functions.NewFunctionsFetchFunctionHistoryParams()
		p = p.WithStackID(args[0]).WithSha(":latest").WithFunction(args[1])
		resp, err := client.Functions.FunctionsFetchFunctionHistory(p, getAuth())
		if err != nil {
			apiError(err)
			return err
		}
		log.Debugf("got response: %v", resp)
		var stackfunctionsHistory []*apimodel.StorageFunctionRunInfoListItem

		for _, st := range resp.GetPayload().Runs {
			stackfunctionsHistory = append(stackfunctionsHistory, st)
		}
		all = append(stackfunctionsHistory, all...)

	}
	format := "table"
	displayStackFunctionsHistory(all, format)

	return nil
}
