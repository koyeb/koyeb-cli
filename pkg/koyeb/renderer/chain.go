package renderer

type ChainRenderer struct {
	Renderer
	isFirst bool // true if the current call to Render is the first one
}

// Most of the CLI display only one resource, but some commands need to display
// multiple resources, like the `koyeb deployments describe` command.
//
// The ChainRenderer complies with the Renderer interface and can be used to display multiple resources, like this:
//
//	chain := renderer.NewChainRenderer(<renderer>)
//	chain.Render(resource1)
//	chain.Render(resource2)
//	chain.Render(resource3)
func NewChainRenderer(Base Renderer) *ChainRenderer {
	return &ChainRenderer{
		Renderer: Base,
		isFirst:  true,
	}
}

func (r *ChainRenderer) Render(item ApiResources) *ChainRenderer {
	// This is not the first call to Render, display a separator: a newline for table, nothing for JSON, and `---` for YAML.
	if !r.isFirst {
		r.RenderSeparator()
	}

	if table, ok := r.Renderer.(*TableRenderer); ok {
		table.RenderTitle(item)
	}

	r.isFirst = false

	// Call the actual renderer
	r.Renderer.Render(item)
	return r
}
