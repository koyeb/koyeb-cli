package flags_list

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

type FlagRoute struct {
	BaseFlag
	path string
	port int64
}

func NewRouteListFromFlags(values []string) ([]Flag[koyeb.DeploymentRoute], error) {
	ret := make([]Flag[koyeb.DeploymentRoute], 0, len(values))

	for _, value := range values {
		route := &FlagRoute{BaseFlag: BaseFlag{cliValue: value}}

		if strings.HasPrefix(value, "!") {
			route.markedForDeletion = true
			value = value[1:]
		}

		split := strings.Split(value, ":")
		route.path = split[0]

		if route.markedForDeletion {
			if len(split) > 1 || route.path == "" {
				return nil, &errors.CLIError{
					What: "Error while configuring the service",
					Why:  fmt.Sprintf("unable to parse the route \"%s\"", route.cliValue),
					Additional: []string{
						"To remove a route from the service, prefix it with '!', e.g. '!/' or '!/foo'",
						"The port should not be specified when removing a route from the service",
					},
					Orig:     nil,
					Solution: "Fix the route and try again",
				}
			}
		} else {
			route.port = 80
			if len(split) > 1 {
				portNum, err := strconv.Atoi(split[1])
				if err != nil {
					return nil, &errors.CLIError{
						What: "Error while configuring the service",
						Why:  fmt.Sprintf("unable to parse the port from the route \"%s\"", route.cliValue),
						Additional: []string{
							"Routes must be specified as PATH[:PORT]",
							"PATH is the route to expose (e.g. / or /foo)",
							"PORT must be a valid port number configured with the --ports flag. It can be omitted, in which case it defaults to \"80\"",
						},
						Orig:     nil,
						Solution: "Fix the route and try again",
					}
				}
				route.port = int64(portNum)
			}
		}
		ret = append(ret, route)
	}
	return ret, nil
}

func (f *FlagRoute) IsEqualTo(route koyeb.DeploymentRoute) bool {
	return f.path == *route.Path
}

func (f *FlagRoute) UpdateItem(route *koyeb.DeploymentRoute) {
	route.Path = koyeb.PtrString(f.path)
	route.Port = koyeb.PtrInt64(f.port)
}

func (f *FlagRoute) CreateNewItem() *koyeb.DeploymentRoute {
	item := koyeb.NewDeploymentRouteWithDefaults()
	f.UpdateItem(item)
	return item
}
