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

	"github.com/ghodss/yaml"
	"github.com/go-openapi/runtime"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
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

type ApiResources interface {
	GetTable() TableInfo
	MarshalBinary() ([]byte, error)
}

func render(defaultFormat string, items ...ApiResources) {
	format := defaultFormat
	if outputFormat != "" {
		format = outputFormat
	}

	for idx, item := range items {
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
			if idx < len(items)-1 {
				fmt.Printf("---\n")
			}
		case "json":
			buf, err := item.MarshalBinary()
			if err != nil {
				er(err)
			}
			fmt.Println(string(buf))
		case "table":
			tableInfo := item.GetTable()
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader(tableInfo.headers)
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
			table.AppendBulk(tableInfo.fields)
			table.Render()
		default:
			er("Invalid format")
		}
	}
}

func getField(item interface{}, field string) string {
	// Ugly but simple and generic
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
			// TODO we should format depending of the type
			return fmt.Sprintf("%s", reflect.Indirect(val).FieldByName(t.Field(i).Name))
		}

		spl := strings.Split(field, ".")
		if spl[0] == fieldName && len(spl) > 1 {
			return getField(reflect.Indirect(reflect.Indirect(val).FieldByName(t.Field(i).Name)).Interface(), strings.Join(spl[1:], "."))
			// []
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
	// case CommonErrorInterface:
	//   log.Debug(er)
	//   payload := er.GetPayload()
	//   errorFn("%s: status:%d code:%s", payload.Message, payload.Status, payload.Code)
	// case CommonErrorWithFieldInterface:
	//   log.Debug(er)
	//   payload := er.GetPayload()
	//   for _, f := range payload.Fields {
	//     log.Errorf("Error on field %s: %s", f.Field, f.Description)
	//   }
	//   errorFn("%s: status:%d code:%s", payload.Message, payload.Status, payload.Code)
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
		// Workarround for https://github.com/go-swagger/go-swagger/issues/1929
		if strings.Contains(err.Error(), "is not supported by the TextConsumer, can be resolved by supporting TextUnmarshaler interface") {
			log.Debug(err)
			errorFn("Unable to process server response")
		} else {
			log.Debugf("Unhandled %T error: %v", err, err)
			errorFn("%v", err)
		}
	}
}
