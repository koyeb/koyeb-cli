package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *ProjectHandler) Describe(ctx *CLIContext, cmd *cobra.Command, args []string) error {
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
	ctx.Renderer.Render(NewDescribeProjectReply(res, full))
	return nil
}

type DescribeProjectReply struct {
	value *koyeb.GetProjectReply
	full  bool
}

func NewDescribeProjectReply(value *koyeb.GetProjectReply, full bool) *DescribeProjectReply {
	return &DescribeProjectReply{
		value: value,
		full:  full,
	}
}

func (DescribeProjectReply) Title() string {
	return "Project"
}

func (r *DescribeProjectReply) MarshalBinary() ([]byte, error) {
	return r.value.GetProject().MarshalJSON()
}

func (r *DescribeProjectReply) Headers() []string {
	return []string{"id", "name", "description", "created_at", "updated_at"}
}

func (r *DescribeProjectReply) Fields() []map[string]string {
	item := r.value.GetProject()
	return []map[string]string{map[string]string{
		"id":          renderer.FormatID(item.GetId(), r.full),
		"name":        item.GetName(),
		"description": item.GetDescription(),
		"created_at":  renderer.FormatTime(item.GetCreatedAt()),
		"updated_at":  renderer.FormatTime(item.GetUpdatedAt()),
	}}
}
