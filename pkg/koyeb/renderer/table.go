package renderer

import (
	"fmt"
	"os"

	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
)

type TableRenderer struct{}

func (r *TableRenderer) Render(item ApiResources) {
	table := tablewriter.NewWriter(os.Stdout)
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

	// Only render the header if there is more than one column
	header := item.Headers()
	if len(header) > 1 {
		table.SetHeader(header)
	}

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
}

func (r *TableRenderer) RenderTitle(item ApiResources) {
	fmt.Println(aurora.Bold(item.Title()))
}

func (r *TableRenderer) RenderSeparator() {
	fmt.Println("")
}
