package renderer

type ChainRenderer struct {
	Renderer
	err     error
	isFirst bool
}

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

	// If the renderer is a table, display the title and a separator (except for the first row)
	if table, ok := r.Renderer.(*TableRenderer); ok {
		if !r.isFirst {
			table.RenderSeparator()
		}
		table.RenderTitle(item)
	}
	r.isFirst = false

	// Call the actual renderer
	r.err = r.Renderer.Render(item)
	return r
}

func (r *ChainRenderer) Err() error {
	return r.err
}
