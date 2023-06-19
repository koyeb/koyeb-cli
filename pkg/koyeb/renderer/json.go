package renderer

import "fmt"

type JSONRenderer struct{}

func (r *JSONRenderer) Render(item ApiResources) error {
	buf, err := item.MarshalBinary()
	if err != nil {
		return err
	}
	fmt.Println(string(buf))
	return nil
}
