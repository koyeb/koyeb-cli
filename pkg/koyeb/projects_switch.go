package koyeb

import (
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *ProjectHandler) Switch(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	project, err := ResolveProjectArgs(ctx, args[0])
	if err != nil {
		return err
	}

	if err := SwitchProjectConfig(project); err != nil {
		return &errors.CLIError{
			What: "Unable to switch the current project",
			Why:  "we were unable to write the configuration file",
			Additional: []string{
				"The command `koyeb project switch` needs to update your configuration file, usually located in $HOME/.koyeb.yaml",
				"If you do not have write access to this file, you can use the --config flag to specify a different location.",
				"Alternatively, you can manually edit the configuration file and set the project field to the project ID you want to use.",
				"You can also provide the project name or UUID with the --project flag.",
			},
			Orig:     err,
			Solution: "Fix the issue preventing the CLI to write the configuration file, or manually edit the configuration file",
		}
	}

	return nil
}
