package renderer

import "fmt"

type JSONRenderer struct{}

func (r *JSONRenderer) Render(item ApiResources) {
	buf, err := item.MarshalBinary()
	// Should never happen, since all the fields of item are marshable
	if err != nil {
		panic("Unable to marshal resource")
	}
	fmt.Println(string(buf))
}

func (r *JSONRenderer) RenderSeparator() {
}
