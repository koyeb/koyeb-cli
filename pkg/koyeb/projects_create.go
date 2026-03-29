package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

func (h *ProjectHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string, createProject *koyeb.CreateProject) error {
	createProject.SetName(args[0])

	res, err := h.CreateProject(ctx, createProject)
	if err != nil {
		return err
	}

	full := GetBoolFlags(cmd, "full")
	getProjectsReply := NewGetProjectReply(ctx.Mapper, &koyeb.GetProjectReply{Project: res.Project}, full)
	ctx.Renderer.Render(getProjectsReply)
	return nil
}
