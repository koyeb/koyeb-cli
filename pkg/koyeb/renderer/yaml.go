package renderer

import (
	"fmt"

	"github.com/ghodss/yaml"
)

type YAMLRenderer struct{}

func (r *YAMLRenderer) Render(item ApiResources) {
	buf, err := item.MarshalBinary()
	// Should never happen, since all the fields of item are marshable
	if err != nil {
		panic("Unable to marshal resource")
	}
	y, err := yaml.JSONToYAML(buf)
	if err != nil {
		panic("Unable to convert JSON to YAML")
	}
	fmt.Printf("%s", string(y))
}

func (r *YAMLRenderer) RenderSeparator() {
	fmt.Println("---")
}
