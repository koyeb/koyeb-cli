package koyeb

import (
	"fmt"
	"net/url"
	"reflect"

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
