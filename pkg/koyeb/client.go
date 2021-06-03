package koyeb

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/go-openapi/runtime"
	"github.com/iancoleman/strcase"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func getApiClient() *koyeb.APIClient {
	u, err := url.Parse(apiurl)
	if err != nil {
		er(err)
	}

	log.Debugf("Using host: %s using %s", u.Host, u.Scheme)

	config := koyeb.NewConfiguration()
	config.Servers[0].URL = u.String()
	config.Debug = debug

	return koyeb.NewAPIClient(config)
}

func getAuth(ctx context.Context) context.Context {
	return context.WithValue(ctx, koyeb.ContextAccessToken, token)
}

type UpdateApiResources interface {
	New() interface{}
	Append(interface{})
}

type TableInfo struct {
	headers []string
	fields  [][]string
}

type WithTitle interface {
	Title() string
}

type ApiResources interface {
	Headers() []string
	Fields() []map[string]string
	MarshalBinary() ([]byte, error)
}

func render(defaultFormat string, items interface{}) {
	format := defaultFormat
	if outputFormat != "" {
		format = outputFormat
	}

	var table *tablewriter.Table
	if format == "table" || format == "detail" {
		table = tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(true)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.SetTablePadding("\t")
		table.SetNoWhiteSpace(true)
	}

	all, ok := items.([]ApiResources)
	if !ok {
		log.Fatalf("Invalid item type %T", items)
	}

	for idx, item := range all {

		switch format {
		case "yaml":
			buf, err := item.MarshalBinary()
			if err != nil {
				er(err)
			}
			y, err := yaml.JSONToYAML(buf)
			if err != nil {
				fmt.Printf("err: %v\n", err)
				return
			}
			fmt.Printf("%s", string(y))
			if idx < len(all)-1 {
				fmt.Printf("---\n")
			}
		case "json":
			buf, err := item.MarshalBinary()
			if err != nil {
				er(err)
			}
			fmt.Println(string(buf))
		case "detail":
			if title, ok := item.(WithTitle); ok {
				fmt.Println(aurora.Bold(title.Title()))
			}
			fields := [][]string{}
			for _, field := range item.Fields() {
				for _, h := range item.Headers() {
					fields = append(fields, append([]string{h}, field[h]))
				}
			}
			table.AppendBulk(fields)
		case "table":
			table.SetHeader(item.Headers())
			fields := [][]string{}
			for _, field := range item.Fields() {
				current := []string{}
				for _, h := range item.Headers() {
					current = append(current, field[h])
				}
				fields = append(fields, current)
			}
			table.AppendBulk(fields)
		default:
			er("Invalid format")
		}

	}

	if format == "table" || format == "detail" {
		table.Render()
	}
}

func GetField(item interface{}, field string) string {
	val := reflect.ValueOf(item)
	t := val.Type()
	for i := 0; i < t.NumField(); i++ {
		fieldName := ""
		if jsonTag := t.Field(i).Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			if commaIdx := strings.Index(jsonTag, ","); commaIdx > 0 {
				fieldName = jsonTag[:commaIdx]
			}
		}

		if fieldName == field {
			f := reflect.Indirect(reflect.Indirect(val).FieldByName(t.Field(i).Name))
			switch val := f.Interface().(type) {
			case string:
				return fmt.Sprintf("%s", val)
			case []koyeb.Domain:
				ret := []string{}
				for _, d := range val {
					ret = append(ret, fmt.Sprintf("%s", d.GetName()))
				}
				return strings.Join(ret, " ")
			case time.Time:
				return fmt.Sprintf("%s", val)
			default:
				log.Debugf("type not supported %T %v", val, val)
				return fmt.Sprintf("%s", val)
			}
		}

		spl := strings.Split(field, ".")
		if spl[0] == fieldName && len(spl) > 1 {
			return GetField(reflect.Indirect(reflect.Indirect(val).FieldByName(t.Field(i).Name)).Interface(), strings.Join(spl[1:], "."))
		}

	}
	return "<empty>"
}

type CommonErrorInterface interface {
}
type CommonErrorWithFieldInterface interface {
}

func logApiError(err error) {
	renderApiError(err, log.Errorf)
}

func fatalApiError(err error) {
	renderApiError(err, log.Fatalf)
}

func renderApiError(err error, errorFn func(string, ...interface{})) {

	switch er := err.(type) {
	case koyeb.GenericOpenAPIError:
		switch mod := er.Model().(type) {
		case koyeb.ErrorWithFields:
			for _, f := range mod.GetFields() {
				log.Errorf("Error on field %s: %s", f.GetField(), f.GetDescription())
			}
			errorFn("%s: status:%d code:%s", mod.GetMessage(), mod.GetStatus(), mod.GetCode())
		case koyeb.Error:
			errorFn("%s: status:%d code:%s", mod.GetMessage(), mod.GetStatus(), mod.GetCode())
		default:
			errorFn("Unhandled error %T: %s", mod, er.Error())
		}
	case *runtime.APIError:
		e := er.Response.(runtime.ClientResponse)
		if debug {
			respBody := e.Body()
			body, _ := ioutil.ReadAll(respBody)
			log.Debugf("%s", body)
		}
		errorFn("%s", e.Message())
	case *json.UnmarshalTypeError:
		log.Debug(err)
		errorFn("Unable to process server response")
	default:
		log.Debugf("Unhandled %T error: %v", err, err)
		errorFn("%v", err)
	}
}

func SyncFlags(cmd *cobra.Command, args []string, i interface{}) {
	cmd.LocalFlags().VisitAll(
		func(flag *pflag.Flag) {
			if !flag.Changed {
				return
			}
			funcName := fmt.Sprintf("Set%s", strcase.ToCamel(flag.Name))
			meth := reflect.ValueOf(i).MethodByName(funcName)
			if !meth.IsValid() {
				log.Debugf("Unable to find setter %s on %T\n", funcName, i)
				return
			}
			meth.Call([]reflect.Value{reflect.ValueOf(flag.Value.String())})
		})
}
