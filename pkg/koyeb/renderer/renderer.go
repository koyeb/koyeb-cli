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

type Renderable interface {
	Render(string) error
}

type MultiRenderer struct {
	renderers []Renderable
}

func NewMultiRenderer(renderers ...Renderable) Renderable {
	return &MultiRenderer{renderers: renderers}
}

func (r *MultiRenderer) Render(format string) error {
	for _, f := range r.renderers {
		err := f.Render(format)
		if err != nil {
			return err
		}
	}
	return nil
}

type SeparatorRenderer struct {
}

func NewSeparatorRenderer() Renderable {
	return &SeparatorRenderer{}
}

func (r *SeparatorRenderer) Render(format string) error {
	if format == "" {
		fmt.Println("")
	}
	return nil
}

type TitleRenderer struct {
	withTitle WithTitle
}

func NewTitleRenderer(withTitle WithTitle) Renderable {
	return &TitleRenderer{withTitle: withTitle}
}

func (r *TitleRenderer) Render(format string) error {
	if format == "" {
		fmt.Println(aurora.Bold(r.withTitle.Title()))
	}
	return nil
}
