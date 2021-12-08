package koyeb

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func getAuth(ctx context.Context) context.Context {
	return context.WithValue(ctx, koyeb.ContextAccessToken, token)
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
