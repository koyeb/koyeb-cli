package flags_list

type FlagAddon struct {
	BaseFlag
}

func NewAddonsListFromFlags(values []string) ([]Flag[string], error) {
	ret := make([]Flag[string], 0, len(values))

	for _, value := range values {
		addon := &FlagAddon{BaseFlag: BaseFlag{cliValue: value}}

		ret = append(ret, addon)
	}
	return ret, nil
}

func (f *FlagAddon) IsEqualTo(item string) bool {
	return f.cliValue == item
}

func (f *FlagAddon) UpdateItem(item *string) {
}

func (f *FlagAddon) CreateNewItem() *string {
	return &f.cliValue
}
