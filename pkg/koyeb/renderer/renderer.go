package renderer

import "fmt"

type WithTitle interface {
	Title() string
}

type ApiResources interface {
	Headers() []string
	Fields() []map[string]string
	MarshalBinary() ([]byte, error)
}

func MultiRenderer(funcs ...func() error) error {
	for i, f := range funcs {
		if i > 0 && i <= len(funcs) {
			fmt.Println("")
		}
		err := f()
		if err != nil {
			return err
		}
	}
	return nil
}
