package koyeb

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"

	apiclient "github.com/koyeb/koyeb-cli/pkg/kclient/client"
	apimodel "github.com/koyeb/koyeb-cli/pkg/kclient/models"
)

func getApiClient() *apiclient.AccountAccountProto {
	u, err := url.Parse(apiurl)
	if err != nil {
		er(err)
	}

	log.Debugf("Using host: %s using %s", u.Host, u.Scheme)
	transport := httptransport.New(u.Host, "", []string{u.Scheme})
	transport.SetDebug(debug)

	return apiclient.New(transport, strfmt.Default)
}

func getAuth() runtime.ClientAuthInfoWriter {
	return httptransport.BearerToken(token)
}

type UpdateApiResources interface {
	New() interface{}
	Append(interface{})
}

type ApiResources interface {
	GetHeaders() []string
	New() interface{}
	GetTableFields() [][]string
	MarshalBinary() ([]byte, error)
}

func render(item ApiResources, defaultFormat string) {
	format := defaultFormat
	if outputFormat != "" {
		format = outputFormat
	}

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
		fmt.Println(string(y))
	case "json":
		buf, err := item.MarshalBinary()
		if err != nil {
			er(err)
		}
		fmt.Println(string(buf))
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader(item.GetHeaders())
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
		table.AppendBulk(item.GetTableFields())
		table.Render()
	default:
		er("Invalid format")
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
	GetPayload() *apimodel.CommonError
}
type CommonErrorWithFieldInterface interface {
	GetPayload() *apimodel.CommonErrorWithFields
}

func apiError(err error) {

	switch er := err.(type) {
	case CommonErrorInterface:
		log.Debug(er)
		payload := er.GetPayload()
		log.Errorf("%s: status:%d code:%s", payload.Message, payload.Status, payload.Code)
	case CommonErrorWithFieldInterface:
		log.Debug(er)
		payload := er.GetPayload()
		log.Errorf("%s: status:%d code:%s", payload.Message, payload.Status, payload.Code)
		for _, f := range payload.Fields {
			log.Errorf("Error on field %s: %s", f.Field, f.Description)
		}
	case *runtime.APIError:
		e := er.Response.(runtime.ClientResponse)
		if debug {
			respBody := e.Body()
			body, _ := ioutil.ReadAll(respBody)
			log.Errorf("%s", body)
		}
	default:
		log.Error(err)
	}

	er, ok := err.(*runtime.APIError)
	if ok {
		e := er.Response.(runtime.ClientResponse)
		if debug {
			respBody := e.Body()
			body, _ := ioutil.ReadAll(respBody)
			log.Errorf("%s", body)
		}
		log.Errorf("%s", e.Message())
	}
}
