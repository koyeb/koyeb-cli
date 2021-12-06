package renderer

import (
	"errors"
	"fmt"
	"os"

	"github.com/ghodss/yaml"
	"github.com/olekukonko/tablewriter"
)

type ListRenderer struct {
	items []ApiResources
}

func NewListRenderer(items ...ApiResources) Renderable {
	return &ListRenderer{items: items}
}

func (r *ListRenderer) Render(format string) error {
	for i, item := range r.items {
		if i > 0 && i <= len(r.items) {
			fmt.Println("")
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
			table.SetHeader(item.Headers())
			fields := [][]string{}
			for _, field := range item.Fields() {
				current := []string{}
				for _, h := range item.Headers() {
					current = append(current, field[h])
				}
				fields = append(fields, current)
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
