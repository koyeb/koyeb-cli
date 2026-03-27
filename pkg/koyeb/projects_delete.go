package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ProjectHandler) Delete(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	project, err := ResolveProjectArgs(ctx, args[0])
	if err != nil {
		return err
	}

	_, resp, err := ctx.Client.ProjectsApi.DeleteProject(ctx.Context, project).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while deleting the project `%s`", args[0]),
			err,
			resp,
		)
	}

	if ctx.Project != "" && ctx.Project == project {
		if err := ClearProjectConfig(); err != nil {
			return &errors.CLIError{
				What: "Unable to clear the current project",
				Why:  "the project was deleted, but we were unable to update the configuration file",
				Additional: []string{
					"The command `koyeb project delete` needs to update your configuration file, usually located in $HOME/.koyeb.yaml",
					"Alternatively, you can manually edit the configuration file and clear the project field.",
				},
				Orig:     err,
				Solution: "Fix the issue preventing the CLI to write the configuration file, or manually edit the configuration file",
			}
		}
	}

	log.Infof("Project %s deleted.", args[0])
	return nil
}
