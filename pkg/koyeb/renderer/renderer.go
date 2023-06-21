// Package renderer provides a set of renderers to display API resources.
//
// The default TableRenderer displays the resources as a table. The JSONRenderer
// and YAMLRenderer display the resources as JSON and YAML respectively.
// ChainRenderer can be used to display multiple resources.
//
// The resource to display must implement the ApiResources interface.
package renderer

type ApiResources interface {
	Headers() []string
	Fields() []map[string]string
	MarshalBinary() ([]byte, error)
	Title() string
}

type Renderer interface {
	Render(ApiResources) error
	RenderSeparator()
}

func NewRenderer(format string) Renderer {
	// XXX jcastets: use an enum instead of a string, requires to parse the flag as an enum
	switch format {
	case "json":
		return &JSONRenderer{}
	case "yaml":
		return &YAMLRenderer{}
	default:
		return &TableRenderer{}
	}
}
