package koyeb

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"regexp"

	"github.com/iancoleman/strcase"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var authorizationHeaderRegexp = regexp.MustCompile("(?m)^Authorization:.*$")

// DebugTransport overrides the default HTTP transport to log the request and the response using our logger.
type DebugTransport struct {
	http.RoundTripper
}

func (t *DebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if reqData, dumpErr := httputil.DumpRequestOut(req, true); dumpErr == nil {
		var safeReqData string

		// Hide the token in the Authorization header
		if !debugFull {
			safeReqData = authorizationHeaderRegexp.ReplaceAllString(string(reqData), "Authorization: <HIDDEN, add --debug-full to show the token>")
		} else {
			safeReqData = string(reqData)
		}
		log.Debug(fmt.Sprintf("========== HTTP request ==========\n%s\n========== end of request ==========\n", safeReqData))
	}

	resp, err := t.RoundTripper.RoundTrip(req)

	if respData, dumpErr := httputil.DumpResponse(resp, true); dumpErr == nil {
		log.Debug(fmt.Sprintf("========== HTTP response ==========\n%s\n========== end of response ==========\n", respData))
	}
	return resp, err
}

func getApiClient() (*koyeb.APIClient, error) {
	u, err := url.Parse(apiurl)
	if err != nil {
		return nil, err
	}

	log.Debugf("Using host: %s using %s", u.Host, u.Scheme)

	config := koyeb.NewConfiguration()
	config.Servers[0].URL = u.String()
	config.HTTPClient = &http.Client{
		Transport: &DebugTransport{http.DefaultTransport},
	}

	return koyeb.NewAPIClient(config), nil
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
