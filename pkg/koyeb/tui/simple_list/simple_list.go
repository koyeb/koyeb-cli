package simple_list

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

var (
	prefixStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	promptStyle   = lipgloss.NewStyle().Bold(true)
	responseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	abortStyle    = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("9"))
	errorStyle    = lipgloss.NewStyle().Background(lipgloss.Color("1")).Foreground(lipgloss.Color("15")) // Red background, white foreground
	successStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))                       // Green
)

type SimpleListItem struct {
	Id          string
	Name        string
	Description string
}

// This method needs to be implemented to satisfy the list.Item interface, even
// though we disable filtering in NewSimpleList.
func (item SimpleListItem) FilterValue() string {
	return ""
}

type SimpleList struct {
	// The prompt to display before the list.
	prompt string
	// The underlying list model.
	list list.Model
	// Set to true when the user presses Ctrl+C.
	abort bool
	// The function that will be called when the user submits the input. Set by SetSubmitHandler.
	onSubmit func(SimpleListItem) tea.Msg
	// Error returned by the submit handler.
	submitError error
	// Set to true when the submit handler returns a SubmitOkMsg.
	submitSuccess bool
}

func New(choices []SimpleListItem) SimpleList {
	model := SimpleList{
		onSubmit: func(SimpleListItem) tea.Msg { return nil },
	}

	items := []list.Item{}
	for _, choice := range choices {
		items = append(items, choice)
	}

	const defaultWidth = 40
	model.list = list.New(items, SimpleListItemDelegate{}, defaultWidth, len(items)+1)
	model.list.SetShowTitle(false)
	model.list.SetShowStatusBar(false)
	model.list.SetFilteringEnabled(false)
	model.list.SetShowHelp(false)
	model.list.SetShowPagination(false)
	// Go back to the first item after reaching the end of the list
	model.list.InfiniteScrolling = true
	// By default, model.list quits if `q` or `ctrl+c` is pressed. Disable this behavior
	model.list.DisableQuitKeybindings()
	return model
}

func (model *SimpleList) SetPrompt(prompt string) {
	model.prompt = prompt
}

func (model *SimpleList) SetSubmitHandler(handler func(SimpleListItem) tea.Msg) {
	model.onSubmit = handler
}

func (model SimpleList) Init() tea.Cmd {
	return nil
}

func (model *SimpleList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case QuitMsg:
		return model, tea.Quit
	case SubmitOkMsg:
		model.submitSuccess = true
		return model, QuitCmd()
	case SubmitErrorMsg:
		model.submitError = msg.Error
		return model, nil
	case tea.WindowSizeMsg:
		model.list.SetWidth(msg.Width)
		return model, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			model.abort = true
			return model, QuitCmd()
		case tea.KeyEnter:
			selected := model.list.SelectedItem().(SimpleListItem)
			return model, model.SubmitCmd(selected)
		// Reset the error when the user moves the cursor again
		default:
			model.submitError = nil
		}
	}
	var cmd tea.Cmd
	model.list, cmd = model.list.Update(msg)
	return model, cmd
}

func (model SimpleList) View() string {
	out := strings.Builder{}

	out.WriteString(fmt.Sprintf("%s %s", prefixStyle.Render(">>"), promptStyle.Render(model.prompt)))

	if model.abort {
		out.WriteString(fmt.Sprintf("%s\n", abortStyle.Render(" woops, abort")))
	} else if model.submitError != nil {
		out.WriteString("\n")
		out.WriteString(model.list.View())
		out.WriteString("\n")
		out.WriteString(errorStyle.Render(model.submitError.Error()))
	} else if model.submitSuccess {
		selected := model.list.SelectedItem().(SimpleListItem)
		out.WriteString(fmt.Sprintf(" %s %s\n", responseStyle.Render(selected.Name), successStyle.Render("✔")))
	} else {
		out.WriteString("\n")
		out.WriteString(model.list.View())
	}
	return out.String()
}

func (model SimpleList) Execute() (bool, error) {
	program, err := tea.NewProgram(&model).Run()
	if err != nil {
		return true, &errors.CLIError{
			What:     "Unable to initialize list input",
			Why:      "we were unable to setup your terminal for interactive input",
			Orig:     err,
			Solution: "Please try again. If the problem still persists, please open an issue at https://github.com/koyeb/koyeb-cli/issues/new",
		}
	}
	if lastModel := program.(*SimpleList); lastModel.abort {
		return true, nil
	}
	return false, nil
}
