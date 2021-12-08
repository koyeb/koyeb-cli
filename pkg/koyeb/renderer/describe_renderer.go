package renderer

import (
	"errors"
	"fmt"
	"os"

	"github.com/ghodss/yaml"
	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
)

type DescribeItemRenderer struct {
	items []ApiResources
}

func NewDescribeItemRenderer(items ...ApiResources) Renderable {
	return &DescribeItemRenderer{items}
}

func (r *DescribeItemRenderer) Render(format string) error {
	for i, item := range r.items {
		if format == "" || format == "table" {
			if title, ok := item.(WithTitle); ok {
				if i > 0 && i <= len(r.items) {
					fmt.Println("")
				}
				fmt.Println(aurora.Bold(title.Title()))
			}
		}

		switch format {
		case "", "table":
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
		case "yaml":
			buf, err := item.MarshalBinary()
			if err != nil {
				return err
			}
			y, err := yaml.JSONToYAML(buf)
			if err != nil {
				return err
			}
			fmt.Printf("%s", string(y))
		case "json":
			buf, err := item.MarshalBinary()
			if err != nil {
				return err
			}
			fmt.Println(string(buf))
		default:
			return errors.New("Invalid format")
		}
	}
	return nil
}

func NewDescribeRenderer(items ...ApiResources) Renderable {
	renderers := []Renderable{}
	renderers = append(renderers, NewDescribeItemRenderer(items[0]))
	for _, item := range items[1:] {
		renderers = append(renderers, NewSeparatorRenderer())
		if withTitle, ok := item.(WithTitle); ok {
			renderers = append(renderers, NewTitleRenderer(withTitle))
		}
		renderers = append(renderers, NewListRenderer(item))
	}
	return NewMultiRenderer(renderers...)
}
