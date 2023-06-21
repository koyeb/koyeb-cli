package renderer

type ChainRenderer struct {
	Renderer
	err     error // used to store the error returned by the last call to Render
	isFirst bool  // true if the current call to Render is the first one
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
//	chain.Err()
//
// If one of the calls to Render returns an error, the following calls are
// skipped and the error is returned by Err().
func NewChainRenderer(Base Renderer) *ChainRenderer {
	return &ChainRenderer{
		Renderer: Base,
		isFirst:  true,
	}
}

func (r *ChainRenderer) Render(item ApiResources) *ChainRenderer {
	// A former call to Render returned an error: skip the current call
	if r.err != nil {
		return r
	}

	// This is not the first call to Render, display a separator: a newline for table, nothing for JSON, and `---` for YAML.
	if !r.isFirst {
		r.RenderSeparator()
	}

	if table, ok := r.Renderer.(*TableRenderer); ok {
		table.RenderTitle(item)
	}

	r.isFirst = false

	// Call the actual renderer
	r.err = r.Renderer.Render(item)
	return r
}

// Err returns the error returned by the first call to Render that returned an error.
func (r *ChainRenderer) Err() error {
	return r.err
}
