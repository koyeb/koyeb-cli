package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *ProjectHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string, createProject *koyeb.CreateProject) error {
	res, resp, err := ctx.Client.ProjectsApi.CreateProject(ctx.Context).Project(*createProject).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while creating the project `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	ctx.Renderer.Render(NewGetProjectReply(&koyeb.GetProjectReply{Project: res.Project}, full))
	return nil
}
