package flags_list

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

type CheckType int

const (
	HealthCheckHTTP CheckType = iota
	HealthCheckTCP
)

type FlagHealthCheck struct {
	BaseFlag
	port      int64
	checkType CheckType
	path      string // Only used for HTTP healthchecks
}

// Parse the list of values in the form PORT[:TYPE[:PATH]]
func NewHealthcheckListFromFlags(values []string) ([]Flag[koyeb.DeploymentHealthCheck], error) {
	ret := make([]Flag[koyeb.DeploymentHealthCheck], 0, len(values))

	for _, value := range values {
		hc := &FlagHealthCheck{BaseFlag: BaseFlag{cliValue: value}}
		components := strings.Split(value, ":")

		if len(components) > 3 {
			return nil, &errors.CLIError{
				What: "Error while configuring the service",
				Why:  fmt.Sprintf("unable to parse the healthcheck \"%s\"", hc.cliValue),
				Additional: []string{
					"Healtchecks must be specified as PORT[:TYPE[:PATH]]",
					"PORT must be a valid port number (e.g. 80)",
					"TYPE must be either \"http\" or \"tcp\". It can be omitted, in which case it defaults to \"http\"",
					"PATH is the path to check for http checks. It can be omitted, in which case it defaults to \"/\". For tcp checks, PATH is ignored",
					"To remove a healthcheck from the service, prefix it with '!', e.g. '!80'",
				},
				Orig:     nil,
				Solution: "Fix the healthcheck and try again",
			}
		}

		if strings.HasPrefix(components[0], "!") {
			if len(components) > 1 {
				return nil, &errors.CLIError{
					What: "Error while configuring the service",
					Why:  fmt.Sprintf("unable to parse the healthcheck \"%s\"", hc.cliValue),
					Additional: []string{
						"To remove a healthcheck from the service, prefix the port with '!', e.g. '!80'",
						"The healthcheck type and path should not be specified to removing it from the service",
					},
					Orig:     nil,
					Solution: "Fix the healthcheck and try again",
				}
			}

			hc.markedForDeletion = true
			components[0] = components[0][1:]
		}

		// Parse the PORT
		port, err := strconv.Atoi(components[0])
		if err != nil {
			return nil, &errors.CLIError{
				What: "Error while configuring the service",
				Why:  fmt.Sprintf("unable to parse the port from the healthcheck \"%s\"", hc.cliValue),
				Additional: []string{
					"PORT is required and must be a valid port number (e.g. 80 or 443)",
				},
				Orig:     nil,
				Solution: "Fix the port and try again",
			}
		}

		healthCheckType := "http"
		if len(components) >= 2 {
			healthCheckType = components[1]
		}

		switch healthCheckType {
		case "http":
			hc.checkType = HealthCheckHTTP
			hc.port = int64(port)
			hc.path = "/"
			if len(components) == 3 {
				hc.path = components[2]
			}
		case "tcp":
			hc.checkType = HealthCheckTCP
			hc.port = int64(port)
		default:
			return nil, &errors.CLIError{
				What: "Error while configuring the service",
				Why:  fmt.Sprintf("unable to parse the protocol from the check \"%s\"", hc.cliValue),
				Additional: []string{
					"The healthcheck protocol must be either \"http\" or \"tcp\"",
					"It can be omitted, in which case it defaults to \"http\"",
				},
				Orig:     nil,
				Solution: "Fix the healthcheck and try again",
			}
		}
		ret = append(ret, hc)
	}
	return ret, nil
}

// IsEqualTo is called to check if a flag given by the user corresponds to a
// given healthcheck. If the flag is a http healthcheck, e.g. "80:http", we
// should return true even if the healthcheck if a TCP healthcheck, as we want
// to allow the user to change the type of the healthcheck.
func (f *FlagHealthCheck) IsEqualTo(hc koyeb.DeploymentHealthCheck) bool {
	http, ok := hc.GetHttpOk()
	if ok {
		return f.port == *http.Port
	}
	tcp, ok := hc.GetTcpOk()
	if ok {
		return f.port == *tcp.Port
	}
	panic("should never happen - flags are always with a valid check type")
}

func (f *FlagHealthCheck) UpdateItem(hc *koyeb.DeploymentHealthCheck) {
	switch f.checkType {
	case HealthCheckHTTP:
		hc.Tcp = nil // force the healthcheck to be HTTP
		httpHealthCheck := koyeb.NewHTTPHealthCheck()
		httpHealthCheck.Port = koyeb.PtrInt64(int64(f.port))
		httpHealthCheck.Path = koyeb.PtrString(f.path)
		hc.SetHttp(*httpHealthCheck)
	case HealthCheckTCP:
		hc.Http = nil // force the healthcheck to be TCP
		tcpHealthCheck := koyeb.NewTCPHealthCheck()
		tcpHealthCheck.Port = koyeb.PtrInt64(int64(f.port))
		hc.SetTcp(*tcpHealthCheck)
	}
}

func (f *FlagHealthCheck) CreateNewItem() *koyeb.DeploymentHealthCheck {
	item := koyeb.NewDeploymentHealthCheckWithDefaults()
	f.UpdateItem(item)
	return item
}
