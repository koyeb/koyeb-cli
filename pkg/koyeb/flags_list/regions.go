package flags_list

import (
	"strings"
)

type FlagRegion struct {
	BaseFlag
}

func NewRegionsListFromFlags(values []string) ([]Flag[string], error) {
	ret := make([]Flag[string], 0, len(values))

	for _, value := range values {
		region := &FlagRegion{BaseFlag: BaseFlag{cliValue: value}}

		if strings.HasPrefix(value, "!") {
			region.markedForDeletion = true
			region.cliValue = value[1:]
		}
		ret = append(ret, region)
	}
	return ret, nil
}

func (f *FlagRegion) IsEqualTo(item string) bool {
	return f.cliValue == item
}

// UpdateItem does nothing for the region flag. For other flags, eg. --port, this function updates the item from
// the existingItems list. For regions, there is nothing to update: the flag is either --region <name>
// or --region !<name>, there is never a need to update the region.
func (f *FlagRegion) UpdateItem(item *string) {
}

func (f *FlagRegion) CreateNewItem() *string {
	return &f.cliValue
}
