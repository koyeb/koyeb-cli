package flags_list

import (
	"fmt"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

type FlagVolume struct {
	BaseFlag
	volumeId string
	path     string
}

// Parse the list of values in the form VOLUME_NAME:PATH.
// The function is wrapped in another function to allow the caller to provide a function to resolve the volume ID from its name.
func GetNewVolumeListFromFlags(resolveVolumeId func(string) (string, error)) func(values []string) ([]Flag[koyeb.DeploymentVolume], error) {
	return func(values []string) ([]Flag[koyeb.DeploymentVolume], error) {
		ret := make([]Flag[koyeb.DeploymentVolume], 0, len(values))

		for _, value := range values {
			hc := &FlagVolume{BaseFlag: BaseFlag{cliValue: value}}
			components := strings.Split(value, ":")

			if strings.HasPrefix(components[0], "!") {
				if len(components) > 1 {
					return nil, &errors.CLIError{
						What: "Error while configuring the service",
						Why:  fmt.Sprintf("unable to parse the volume \"%s\"", hc.cliValue),
						Additional: []string{
							"To remove a volume from the service, prefix the volume name with '!', e.g. '!myvolume'",
							"The volume mount point should not be specified to removing it from the service",
						},
						Orig:     nil,
						Solution: "Fix the volume and try again",
					}
				}
				volumeId, err := resolveVolumeId(components[0][1:])
				if err != nil {
					return nil, err
				}
				hc.markedForDeletion = true
				hc.volumeId = volumeId
			} else {
				if len(components) != 2 {
					return nil, &errors.CLIError{
						What: "Error while configuring the service",
						Why:  fmt.Sprintf("unable to parse the volume \"%s\"", hc.cliValue),
						Additional: []string{
							"Volumes must be specified as VOLUME_ID:PATH",
							"To remove a volume from the service, prefix it with '!', e.g. '!myvolume'",
						},
						Orig:     nil,
						Solution: "Fix the volume and try again",
					}
				}
				volumeId, err := resolveVolumeId(components[0])
				if err != nil {
					return nil, err
				}
				hc.volumeId = volumeId
				hc.path = components[1]
			}
			ret = append(ret, hc)
		}
		return ret, nil
	}
}

// IsEqualTo is called to check if a flag given by the user corresponds to a
// given voolume. If the flag is a volume, e.g. "myvolume:/data", we
// should return true even if the volume has a different path, as we want
// to allow the user to change the path where a volume is mounted.
func (f *FlagVolume) IsEqualTo(hc koyeb.DeploymentVolume) bool {
	return hc.GetId() == f.volumeId
}

func (f *FlagVolume) UpdateItem(hc *koyeb.DeploymentVolume) {
	hc.Id = &f.volumeId
	hc.Path = &f.path
}

func (f *FlagVolume) CreateNewItem() *koyeb.DeploymentVolume {
	item := koyeb.NewDeploymentVolumeWithDefaults()
	f.UpdateItem(item)
	return item
}
