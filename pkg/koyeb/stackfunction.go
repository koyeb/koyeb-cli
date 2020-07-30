package koyeb

import (
	"fmt"

	"github.com/go-openapi/swag"
	// "github.com/jinzhu/copier"
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
		// RunE:    getStackFunctions,
	}
	runStackFunctionCommand = &cobra.Command{
		Use:     "functions [stack] [name]",
		Aliases: []string{"function"},
		Short:   "Run stack functions",
		Args:    cobra.MinimumNArgs(1),
		// RunE:    getStackFunctions,
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