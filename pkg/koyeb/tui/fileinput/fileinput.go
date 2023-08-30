package fileinput

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/tui/input"
)

type FileInput struct {
	Input input.Input
}

func New() FileInput {
	return FileInput{
		Input: input.New(),
	}
}

func (model FileInput) Init() tea.Cmd {
	return model.Input.Init()
}

func (model *FileInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return model.Input.Update(msg)
}

func (model FileInput) View() string {
	return model.Input.View()
}

func (model FileInput) Execute() (bool, error) {
	program, err := tea.NewProgram(&model).Run()
	if err != nil {
		return true, &errors.CLIError{
			What:     "Unable to initialize file input",
			Why:      "we were unable to setup your terminal for interactive input",
			Orig:     err,
			Solution: "Please try again. If the problem still persists, please open an issue at https://github.com/koyeb/koyeb-cli/issues/new",
		}
	}
	if lastModel := program.(*FileInput); lastModel.Input.IsAborted() {
		return true, nil
	}
	return false, nil
}
