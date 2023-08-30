package input

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type SubmitOkMsg struct{}

type SubmitErrorMsg struct {
	Error error
}

// SubmitCmd returns a tea.Cmd that will call the submit handler with the input
// value. The handler is expected to return SubmitOkMsg or SubmitErrorMsg.
func (model Input) SubmitCmd(value string) tea.Cmd {
	return func() tea.Msg {
		value = strings.TrimSpace(value)
		return model.onSubmit(value)
	}
}

type QuitMsg struct{}

// To exit the program, rather than returning tea.Quit, we return QuitCmd which
// allows us to redraw the model one last time in View().
func QuitCmd() tea.Cmd {
	return func() tea.Msg {
		return QuitMsg{}
	}
}