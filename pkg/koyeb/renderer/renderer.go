package renderer

import (
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/logrusorgru/aurora"
)

type WithTitle interface {
	Title() string
}

type WithMarshal interface {
	MarshalJSON() ([]byte, error)
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

type GenericRenderer struct {
	res   WithMarshal
	title string
}

func NewGenericRenderer(title string, res WithMarshal) *GenericRenderer {
	return &GenericRenderer{
		res:   res,
		title: title,
	}
}

func (a *GenericRenderer) Render(format string) error {
	if format == "" {
		fmt.Println(aurora.Bold(a.title))
		buf, err := a.res.MarshalJSON()
		if err != nil {
			return err
		}
		y, err := yaml.JSONToYAML(buf)
		if err != nil {
			return err
		}
		fmt.Printf("%s", string(y))
	}
	return nil
}
