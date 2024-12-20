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

func GetNewFilestListFromFlags() func(values []string) ([]Flag[koyeb.File], error) {
	return func(values []string) ([]Flag[koyeb.File], error) {
		ret := make([]Flag[koyeb.File], 0, len(values))

		for _, value := range values {
			hc := &FlagFile{BaseFlag: BaseFlag{cliValue: value}}
			components := strings.Split(value, ":")

			if strings.HasPrefix(components[0], "!") {
				if len(components) > 1 {
					return nil, &kerrors.CLIError{
						What: "Error while configuring the service",
						Why:  fmt.Sprintf("unable to parse the file flag value \"%s\"", hc.cliValue),
						Additional: []string{
							"To remove a mounted file from the service, prefix the path with '!', e.g. '!path'",
							"The source should not be specified to remove it from the service",
						},
						Orig:     nil,
						Solution: "Fix the file flag value and try again",
					}
				}
				hc.markedForDeletion = true
				hc.path = components[0][1:]
			} else {
				if len(components) != 2 && len(components) != 3 {
					return nil, &kerrors.CLIError{
						What: "Error while configuring the service",
						Why:  fmt.Sprintf("unable to parse the file flag value \"%s\"", hc.cliValue),
						Additional: []string{
							"File flag value must be specified as SOURCE:PATH[:PERMISSIONS]",
							"To remove a mounted file from the service, prefix the path with '!', e.g. '!path'",
						},
						Orig:     nil,
						Solution: "Fix the file flag value and try again",
					}
				}
				hc.path = components[1]
				path := components[0]
				if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
					return nil, &kerrors.CLIError{
						What: "Error while configuring the service",
						Why:  fmt.Sprintf("unable to parse the file flag value \"%s\"", hc.cliValue),
						Additional: []string{
							"File flag value must be specified as SOURCE:PATH[:PERMISSIONS]",
							"To remove a mounted file from the service, prefix the path with '!', e.g. '!path'",
						},
						Orig:     nil,
						Solution: "Fix the file flag value and try again",
					}
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return nil, &kerrors.CLIError{
						What: "Error while configuring the service",
						Why:  fmt.Sprintf("unable to read the file flag value \"%s\"", hc.cliValue),
						Additional: []string{
							"File flag value must be specified as SOURCE:PATH[:PERMISSIONS]",
							"To remove a file mount from the service, prefix it with '!', e.g. '!path'",
						},
						Orig:     nil,
						Solution: "Fix the file flag value and try again",
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
						Why:  fmt.Sprintf("unable to parse the file flag value \"%s\"", hc.cliValue),
						Additional: []string{
							"File mount permission must be specified as SOURCE:PATH:PERMISSIONS",
							"To remove a file mount from the service, prefix it with '!', e.g. '!path'",
						},
						Orig:     nil,
						Solution: "Fix the permissions in file flag value and try again",
					}
				}
				hc.permissions = permissions
			}
			ret = append(ret, hc)
		}
		return ret, nil
	}
}

func (f *FlagFile) IsEqualTo(hc koyeb.File) bool {
	return hc.GetPath() == f.path
}

func (f *FlagFile) UpdateItem(hc *koyeb.File) {
	hc.Content = &f.content
	hc.Path = &f.path
	hc.Permissions = &f.permissions
}

func (f *FlagFile) CreateNewItem() *koyeb.File{
	item := koyeb.NewFileWithDefaults()
	f.UpdateItem(item)
	return item
}
