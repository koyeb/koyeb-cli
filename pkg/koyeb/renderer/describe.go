package renderer

import (
	"fmt"
	"os"

	"github.com/ghodss/yaml"
	"github.com/olekukonko/tablewriter"
)

type DescribeItemRenderer interface {
	Render(items ...ApiResources) error
}

func NewDescribeItemRenderer(format string) DescribeItemRenderer {
	switch format {
	case "json":
		return &DescribeItemJSONRenderer{}
	case "yaml":
		return &DescribeItemYAMLRenderer{}
	default:
		return &DescribeItemTableRenderer{}
	}
}

type DescribeItemJSONRenderer struct{}

func (r *DescribeItemJSONRenderer) Render(items ...ApiResources) error {
	for _, item := range items {
		buf, err := item.MarshalBinary()
		if err != nil {
			return err
		}
		fmt.Println(string(buf))
	}
	return nil
}

type DescribeItemYAMLRenderer struct{}

func (r *DescribeItemYAMLRenderer) Render(items ...ApiResources) error {
	for _, item := range items {
		buf, err := item.MarshalBinary()
		if err != nil {
			return err
		}
		y, err := yaml.JSONToYAML(buf)
		if err != nil {
			return err
		}
		fmt.Printf("%s", string(y))
	}
	return nil
}

type DescribeItemTableRenderer struct{}

func (r *DescribeItemTableRenderer) Render(items ...ApiResources) error {
	for i, item := range items {
		fmt.Printf("xxx%v\n", i)
		// if title, ok := item.(WithTitle); ok {
		// 	if i > 0 && i <= len(items) {
		// 		fmt.Println("")
		// 	}
		// 	fmt.Println(aurora.Bold(title.Title()))
		// }

		var table *tablewriter.Table

		table = tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(true)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.SetTablePadding("\t")
		table.SetNoWhiteSpace(true)
		fields := [][]string{}
		for _, field := range item.Fields() {
			for _, h := range item.Headers() {
				fields = append(fields, append([]string{h}, field[h]))
			}
		}
		table.AppendBulk(fields)
		table.Render()
	}
	return nil
}

// func NewDescribeRenderer(items ...ApiResources) Renderable {
// 	renderers := []Renderable{}
// 	renderers = append(renderers, NewDescribeItemRenderer(items[0]))
// 	for _, item := range items[1:] {
// 		renderers = append(renderers, NewSeparatorRenderer())
// 		if withTitle, ok := item.(WithTitle); ok {
// 			renderers = append(renderers, NewTitleRenderer(withTitle))
// 		}
// 		renderers = append(renderers, NewListRenderer(item))
// 	}
// 	return NewMultiRenderer(renderers...)
// }
