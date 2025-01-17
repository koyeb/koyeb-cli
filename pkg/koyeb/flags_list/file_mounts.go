package flags_list

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	kerrors "github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

type FlagFile struct {
	BaseFlag
	path        string
	permissions string
	content     string
}

func GetNewConfigFilestListFromFlags() func(values []string) ([]Flag[koyeb.ConfigFile], error) {
	return func(values []string) ([]Flag[koyeb.ConfigFile], error) {
		ret := make([]Flag[koyeb.ConfigFile], 0, len(values))

		for _, value := range values {
			hc := &FlagFile{BaseFlag: BaseFlag{cliValue: value}}
			components := strings.Split(value, ":")

			if strings.HasPrefix(components[0], "!") {
				if len(components) > 1 {
					return nil, &kerrors.CLIError{
						What: "Error while configuring the service",
						Why:  fmt.Sprintf("unable to parse the config-file flag value \"%s\"", hc.cliValue),
						Additional: []string{
							"To remove a mounted config file from the service, prefix the path with '!', e.g. '!path'",
							"The source should not be specified to remove it from the service",
						},
						Orig:     nil,
						Solution: "Fix the config-file flag value and try again",
					}
				}
				hc.markedForDeletion = true
				hc.path = components[0][1:]
			} else {
				if len(components) != 2 && len(components) != 3 {
					return nil, &kerrors.CLIError{
						What: "Error while configuring the service",
						Why:  fmt.Sprintf("unable to parse the confi-file flag value \"%s\"", hc.cliValue),
						Additional: []string{
							"Config file flag value must be specified as SOURCE:PATH[:PERMISSIONS]",
							"To remove a mounted config file from the service, prefix the path with '!', e.g. '!path'",
						},
						Orig:     nil,
						Solution: "Fix the config-file flag value and try again",
					}
				}
				hc.path = components[1]
				path := components[0]
				if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
					return nil, &kerrors.CLIError{
						What: "Error while configuring the service",
						Why:  fmt.Sprintf(" Unable to locate file at \"%s\"", path),
						Additional: []string{
							"Config file flag value must be specified as SOURCE:PATH[:PERMISSIONS]",
							"To remove a mounted config file from the service, prefix the path with '!', e.g. '!path'",
						},
						Orig:     nil,
						Solution: "Fix the config-file flag value and try again",
					}
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return nil, &kerrors.CLIError{
						What: "Error while configuring the service",
						Why:  fmt.Sprintf("unable to read the file \"%s\"", path),
						Additional: []string{
							"Config file flag value must be specified as SOURCE:PATH[:PERMISSIONS]",
							"To remove a config file mount from the service, prefix it with '!', e.g. '!path'",
						},
						Orig:     nil,
						Solution: "Fix the config-file flag value and try again",
					}
				}
				hc.content = string(data)

				permissions := "0644"
				if len(components) == 3 {
					permissions = components[2]
				}
				if len(permissions) != 4 {
					return nil, &kerrors.CLIError{
						What: "Error while configuring the service",
						Why:  fmt.Sprintf("unable to parse the permissions \"%s\"", permissions),
						Additional: []string{
							"File mount permission must be specified as SOURCE:PATH:PERMISSIONS and in format like 0644",
							"To remove a config file mount from the service, prefix it with '!', e.g. '!path'",
						},
						Orig:     nil,
						Solution: "Fix the permissions in config-file flag value and try again",
					}
				}
				hc.permissions = permissions
			}
			ret = append(ret, hc)
		}
		return ret, nil
	}
}

func (f *FlagFile) IsEqualTo(hc koyeb.ConfigFile) bool {
	return hc.GetPath() == f.path
}

func (f *FlagFile) UpdateItem(hc *koyeb.ConfigFile) {
	hc.Content = &f.content
	hc.Path = &f.path
	hc.Permissions = &f.permissions
}

func (f *FlagFile) CreateNewItem() *koyeb.ConfigFile {
	item := koyeb.NewConfigFileWithDefaults()
	f.UpdateItem(item)
	return item
}
