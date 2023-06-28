// Package renderer provides a set of renderers to display API resources.
//
// The default TableRenderer displays the resources as a table. The JSONRenderer
// and YAMLRenderer display the resources as JSON and YAML respectively.
// ChainRenderer can be used to display multiple resources.
//
// The resource to display must implement the ApiResources interface.
package renderer

import (
	"errors"
)

type ApiResources interface {
	Headers() []string
	Fields() []map[string]string
	MarshalBinary() ([]byte, error)
	Title() string
}

type Renderer interface {
	Render(ApiResources)
	RenderSeparator()
}

// OutputFormat implements the flag.Value interface to parse the --output flag.
type OutputFormat string

const (
	JSONFormat  OutputFormat = "json"
	YAMLFormat  OutputFormat = "yaml"
	TableFormat OutputFormat = "table"
)

func (f *OutputFormat) String() string {
	return string(*f)
}

func (f *OutputFormat) Set(value string) error {
	switch value {
	case "json", "yaml", "table":
		*f = OutputFormat(value)
		return nil
	}
	return errors.New(`invalid output format. Valid values are "json", "yaml" and "table"`)
}

func (f *OutputFormat) Type() string {
	return "output"
}

func NewRenderer(format OutputFormat) Renderer {
	switch format {
	case JSONFormat:
		return &JSONRenderer{}
	case YAMLFormat:
		return &YAMLRenderer{}
	default:
		return &TableRenderer{}
	}
}
