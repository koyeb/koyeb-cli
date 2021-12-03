package renderer

import (
	"errors"
	"fmt"
	"os"

	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
)

func DescribeRenderer(format string, items ...ApiResources) error {
	for i, item := range items {
		if title, ok := item.(WithTitle); ok {
			if i > 0 && i <= len(items) {
				fmt.Println("")
			}
			fmt.Println(aurora.Bold(title.Title()))
		}

		switch format {
		case "":
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
		default:
			return errors.New("Invalid format")
		}
	}
	return nil
}
