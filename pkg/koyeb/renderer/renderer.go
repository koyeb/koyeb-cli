package renderer

type WithTitle interface {
	Title() string
}

type ApiResources interface {
	Headers() []string
	Fields() []map[string]string
	MarshalBinary() ([]byte, error)
}
