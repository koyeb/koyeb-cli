package renderer

import (
	"fmt"

	"github.com/ghodss/yaml"
)

type YAMLRenderer struct{}

func (r *YAMLRenderer) Render(item ApiResources) error {
	buf, err := item.MarshalBinary()
	if err != nil {
		return err
	}
	y, err := yaml.JSONToYAML(buf)
	if err != nil {
		return err
	}
	fmt.Printf("%s", string(y))
	return nil
}
