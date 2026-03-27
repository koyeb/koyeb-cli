package koyeb

import (
	"fmt"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *ProjectHandler) Update(ctx *CLIContext, cmd *cobra.Command, args []string, updateProject *koyeb.Project) error {
	project, err := ResolveProjectArgs(ctx, args[0])
	if err != nil {
		return err
	}

	updateMask := []string{}
	if cmd.Flags().Changed("name") {
		updateMask = append(updateMask, "name")
	}
	if cmd.Flags().Changed("description") {
		updateMask = append(updateMask, "description")
	}

	req := ctx.Client.ProjectsApi.UpdateProject2(ctx.Context, project).Project(*updateProject)
	if len(updateMask) > 0 {
		req = req.UpdateMask(strings.Join(updateMask, ","))
	}

	res, resp, err := req.Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while updating the project `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	ctx.Renderer.Render(NewGetProjectReply(&koyeb.GetProjectReply{Project: res.Project}, full))
	return nil
}
