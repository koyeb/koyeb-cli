package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *ProjectHandler) Get(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	project, err := ResolveProjectArgs(ctx, args[0])
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.ProjectsApi.GetProject(ctx.Context, project).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the project `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	ctx.Renderer.Render(NewGetProjectReply(res, full))
	return nil
}

type GetProjectReply struct {
	value *koyeb.GetProjectReply
	full  bool
}

func NewGetProjectReply(value *koyeb.GetProjectReply, full bool) *GetProjectReply {
	return &GetProjectReply{
		value: value,
		full:  full,
	}
}

func (GetProjectReply) Title() string {
	return "Project"
}

func (r *GetProjectReply) MarshalBinary() ([]byte, error) {
	return r.value.GetProject().MarshalJSON()
}

func (r *GetProjectReply) Headers() []string {
	return []string{"id", "name", "description", "created_at"}
}

func (r *GetProjectReply) Fields() []map[string]string {
	item := r.value.GetProject()
	return []map[string]string{projectFields(item, r.full)}
}

func projectFields(item koyeb.Project, full bool) map[string]string {
	return map[string]string{
		"id":          renderer.FormatID(item.GetId(), full),
		"name":        item.GetName(),
		"description": item.GetDescription(),
		"created_at":  renderer.FormatTime(item.GetCreatedAt()),
	}
}
