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
		if err := updateOrganizationDefaultProjectID(ctx, nil); err != nil {
			return err
		}
	}

	log.Infof("Project %s deleted.", args[0])
	return nil
}
