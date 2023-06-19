package renderer

type ApiResources interface {
	Headers() []string
	Fields() []map[string]string
	MarshalBinary() ([]byte, error)
	Title() string
}

type Renderer interface {
	Render(ApiResources) error
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
