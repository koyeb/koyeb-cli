package simple_list

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	cursorStyle      = lipgloss.NewStyle().Bold(true)
	itemStyle        = lipgloss.NewStyle().Bold(true)
	descriptionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	cursor           = cursorStyle.Render("*")
)

type SimpleListItemDelegate struct{}

func (d SimpleListItemDelegate) Height() int {
	return 1
}

func (d SimpleListItemDelegate) Spacing() int {
	return 0
}

func (d SimpleListItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d SimpleListItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	var str string
	item := listItem.(SimpleListItem)
	if index == m.Index() {
		str = fmt.Sprintf("%s %s %s", cursor, itemStyle.Render(item.Name), descriptionStyle.Render(item.Description))
	} else {
		str = fmt.Sprintf("  %s %s", item.Name, descriptionStyle.Render(item.Description))
	}
	fmt.Fprintf(w, str)
}
