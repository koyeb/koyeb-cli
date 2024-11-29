package flags_list

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	kerrors "github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

type FlagFileMount struct {
	BaseFlag
	raw         *string
	secret      *string
	path        string
	permissions string
}

// Parse the list of values in the form path:secret:name or path:file:<path>
// The function is wrapped in another function to allow the caller to provide a function to resolve the volume ID from its name.
func GetNewFileMountListFromFlags() func(values []string) ([]Flag[koyeb.DeploymentFileMount], error) {
	return func(values []string) ([]Flag[koyeb.DeploymentFileMount], error) {
		ret := make([]Flag[koyeb.DeploymentFileMount], 0, len(values))

		for _, value := range values {
			hc := &FlagFileMount{BaseFlag: BaseFlag{cliValue: value}}
			components := strings.Split(value, ":")

			if strings.HasPrefix(components[0], "!") {
				if len(components) > 1 {
					return nil, &kerrors.CLIError{
						What: "Error while configuring the service",
						Why:  fmt.Sprintf("unable to parse the file mount\"%s\"", hc.cliValue),
						Additional: []string{
							"To remove a file mount from the service, prefix the path with '!', e.g. '!path'",
							"The source should not be specified to remove it from the service",
						},
						Orig:     nil,
						Solution: "Fix the file mount and try again",
					}
				}
				hc.markedForDeletion = true
				hc.path = components[0][1:]
			} else {
				if len(components) != 4 {
					return nil, &kerrors.CLIError{
						What: "Error while configuring the service",
						Why:  fmt.Sprintf("unable to parse the file mount\"%s\"", hc.cliValue),
						Additional: []string{
							"File mounts must be specified as PATH:SOURCE:DATA:PERMISSIONS",
							"To remove a file mount from the service, prefix it with '!', e.g. '!path'",
						},
						Orig:     nil,
						Solution: "Fix the file mount and try again",
					}
				}
				hc.path = components[0]
				source := components[1]
				source = strings.ToLower(source)
				switch source {
				case "secret":
					hc.secret = &components[2]
				case "file":
					path := components[2]
					if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
						return nil, &kerrors.CLIError{
							What: "Error while configuring the service",
							Why:  fmt.Sprintf("unable to parse the file mount\"%s\"", hc.cliValue),
							Additional: []string{
								"File mounts must be specified as PATH:SOURCE:DATA:PERMISSIONS",
								"To remove a file mount from the service, prefix it with '!', e.g. '!path'",
							},
							Orig:     nil,
							Solution: "Fix the file mount and try again",
						}
					}
					data, err := os.ReadFile(components[2])
					if err != nil {
						return nil, &kerrors.CLIError{
							What: "Error while configuring the service",
							Why:  fmt.Sprintf("unable to parse the file mount\"%s\"", hc.cliValue),
							Additional: []string{
								"File mounts must be specified as PATH:SOURCE:DATA:PERMISSIONS",
								"To remove a file mount from the service, prefix it with '!', e.g. '!path'",
							},
							Orig:     nil,
							Solution: "Fix the file mount and try again",
						}
					}
					encoded := base64.URLEncoding.EncodeToString([]byte(data))
					hc.raw = &encoded

				}
				permissions := components[3]
				if len(permissions) != 4 {
					return nil, &kerrors.CLIError{
						What: "Error while configuring the service",
						Why:  fmt.Sprintf("unable to parse the file mount\"%s\"", hc.cliValue),
						Additional: []string{
							"File mounts must be specified as PATH:SOURCE:DATA:PERMISSIONS",
							"To remove a file mount from the service, prefix it with '!', e.g. '!path'",
						},
						Orig:     nil,
						Solution: "Fix the permissions in file mount and try again",
					}
				}
				hc.permissions = permissions
				ret = append(ret, hc)
			}
		}
		return ret, nil
	}
}

func (f *FlagFileMount) IsEqualTo(hc koyeb.DeploymentFileMount) bool {
	return hc.GetPath() == f.path
}

func (f *FlagFileMount) UpdateItem(hc *koyeb.DeploymentFileMount) {
	if f.secret != nil {
		hc.Secret = &koyeb.SecretSource{Name: f.secret}
	}
	if f.raw != nil {
		hc.Raw = &koyeb.RawSource{Content: f.raw}
	}
	hc.Path = &f.path
	hc.Permissions = &f.permissions
}

func (f *FlagFileMount) CreateNewItem() *koyeb.DeploymentFileMount {
	item := koyeb.NewDeploymentFileMountWithDefaults()
	f.UpdateItem(item)
	return item
}
