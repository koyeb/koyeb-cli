package renderer

import (
	"fmt"

	"github.com/logrusorgru/aurora"
)

type WithTitle interface {
	Title() string
}

type ApiResources interface {
	Headers() []string
	Fields() []map[string]string
	MarshalBinary() ([]byte, error)
}

func MultiRenderer(funcs ...func() error) error {
	for _, f := range funcs {
		err := f()
		if err != nil {
			return err
		}
	}
	return nil
}

func SeparatorRenderer(format string) error {
	if format == "" {
		fmt.Println("")
	}
	return nil
}

func TitleRenderer(format string, item ApiResources) error {
	if format == "" {
		if title, ok := item.(WithTitle); ok {
			fmt.Println(aurora.Bold(title.Title()))
		}
	}
	return nil
}
