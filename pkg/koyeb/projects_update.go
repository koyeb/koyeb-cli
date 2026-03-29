package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *ProjectHandler) Update(ctx *CLIContext, cmd *cobra.Command, args []string, updateProject *koyeb.Project) error {
	project, err := h.ResolveProjectArgs(ctx, args[0])
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.ProjectsApi.UpdateProject(ctx.Context, project).Project(*updateProject).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while updating the project `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getProjectsReply := NewGetProjectReply(ctx.Mapper, &koyeb.GetProjectReply{Project: res.Project}, full)
	ctx.Renderer.Render(getProjectsReply)
	return nil
}
