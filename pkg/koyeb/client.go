package koyeb

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"github.com/go-openapi/runtime"
	"github.com/iancoleman/strcase"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
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

type CommonErrorInterface interface {
}
type CommonErrorWithFieldInterface interface {
}

func fatalApiError(err error, resp *http.Response) {
	renderApiError(err, resp, log.Fatalf)
}

type genericError struct {
	Status  int
	Code    string
	Message string
}

func renderHTTPResponse(resp *http.Response) string {
	if resp == nil {
		return "Unhandled error"
	}

	gError := genericError{}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Error status:%d message:unable to read response body", resp.StatusCode)
	}
	err = json.Unmarshal(bodyBytes, &gError)
	if err != nil {
		return fmt.Sprintf("Error status:%d message:%v\n", resp.StatusCode, resp.Body)
	}
	return fmt.Sprintf("Error status:%d code:%s message:%v\n", gError.Status, gError.Code, gError.Message)
}

func renderApiError(err error, resp *http.Response, errorFn func(string, ...interface{})) {

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
		case nil:
			if resp == nil {
				errorFn("Unhandled error %T: %s", err, err.Error())
			} else {
				errorFn(renderHTTPResponse(resp))
			}
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
			if !flag.Changed && flag.DefValue == "" {
				return
			}
			funcName := fmt.Sprintf("Set%s", strcase.ToCamel(flag.Name))
			meth := reflect.ValueOf(i).MethodByName(funcName)
			if !meth.IsValid() {
				log.Debugf("Unable to find setter %s on %T\n", funcName, i)
				return
			}
			switch flag.Value.Type() {
			case "stringSlice":
				v, _ := cmd.LocalFlags().GetStringSlice(flag.Name)
				meth.Call([]reflect.Value{reflect.ValueOf(v)})
			case "intSlice":
				v, _ := cmd.LocalFlags().GetIntSlice(flag.Name)
				meth.Call([]reflect.Value{reflect.ValueOf(v)})
			default:
				meth.Call([]reflect.Value{reflect.ValueOf(flag.Value.String())})
			}
		})
}
