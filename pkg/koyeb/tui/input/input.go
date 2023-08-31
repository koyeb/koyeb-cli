package input

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

var (
	prefixStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")) // Green
	promptStyle  = lipgloss.NewStyle().Bold(true)
	abortStyle   = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("9"))                     // Red
	successStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))                       // Green
	errorStyle   = lipgloss.NewStyle().Background(lipgloss.Color("1")).Foreground(lipgloss.Color("15")) // Red background, white foreground
)

type Input struct {
	// The function that will be called when the user submits the input. Set by SetSubmitHandler.
	onSubmit func(string) tea.Msg
	// The underlying text input model.
	input textinput.Model
	// Set to true when the user presses Ctrl+C.
	abort bool
	// Error returned by the submit handler.
	submitError error
	// Set to true when the submit handler returns a SubmitOkMsg.
	submitSuccess bool
	// Text displayed under the prompt
	Text string
}

func New() Input {
	model := Input{
		input:    textinput.New(),
		onSubmit: func(string) tea.Msg { return nil },
	}
	model.input.Focus()
	return model
}

func (model *Input) SetPrompt(prompt string) {
	prefix := prefixStyle.Render(">>")
	model.input.Prompt = fmt.Sprintf("%s %s ", prefix, promptStyle.Render(prompt))
}

func (model *Input) SetText(text string) {
	model.Text = text
}

func (model *Input) SetSubmitHandler(handler func(string) tea.Msg) {
	model.onSubmit = handler
}

func (model Input) Init() tea.Cmd {
	return textinput.Blink
}

func (model *Input) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case QuitMsg:
		return model, tea.Quit
	case SubmitOkMsg:
		model.input.Blur()
		model.submitSuccess = true
		return model, QuitCmd()
	case SubmitErrorMsg:
		model.submitError = msg.Error
		return model, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			model.input.Blur()
			model.abort = true
			return model, QuitCmd()
		case tea.KeyEnter:
			return model, model.SubmitCmd(model.input.Value())
		// Reset the error when the user starts typing again
		default:
			model.submitError = nil
		}
	}
	var cmd tea.Cmd
	model.input, cmd = model.input.Update(msg)
	return model, cmd
}

func (model Input) View() string {
	out := strings.Builder{}

	out.WriteString(model.input.View())

	if model.submitError != nil {
		out.WriteString("\n")
		out.WriteString(errorStyle.Render(model.submitError.Error()))
	} else if model.abort {
		out.WriteString(abortStyle.Render("woops, abort"))
	} else if model.submitSuccess {
		out.WriteString(successStyle.Render("✔"))
	}
	if model.Text != "" {
		out.WriteString("\n")
		out.WriteString(model.Text)
	}
	return out.String()
}

// Execute prompts the user for input. The boolean returned is true if the user
// aborted the input (Ctrl+C). The error is non-nil if the initialisation of the
// input failed.
func (model Input) Execute() (bool, error) {
	program, err := tea.NewProgram(&model).Run()
	if err != nil {
		return true, &errors.CLIError{
			What:     "Unable to initialize text input",
			Why:      "we were unable to setup your terminal for interactive input",
			Orig:     err,
			Solution: "Please try again. If the problem still persists, please open an issue at https://github.com/koyeb/koyeb-cli/issues/new",
		}
	}
	if lastModel := program.(*Input); lastModel.abort {
		return true, nil
	}
	return false, nil
}

// Returns true if the user aborted the input (Ctrl+C).
func (model Input) IsAborted() bool {
	return model.abort
}
