package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *ProjectHandler) Describe(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	project, err := h.ResolveProjectArgs(ctx, args[0])
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
	describeProjectsReply := NewDescribeProjectReply(ctx.Mapper, res, full)
	ctx.Renderer.Render(describeProjectsReply)
	return nil
}

type DescribeProjectReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetProjectReply
	full   bool
}

func NewDescribeProjectReply(mapper *idmapper.Mapper, value *koyeb.GetProjectReply, full bool) *DescribeProjectReply {
	return &DescribeProjectReply{
		mapper: mapper,
		value:  value,
		full:   full,
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
	fields := map[string]string{
		"id":          renderer.FormatID(item.GetId(), r.full),
		"name":        item.GetName(),
		"description": item.GetDescription(),
		"created_at":  renderer.FormatTime(item.GetCreatedAt()),
		"updated_at":  renderer.FormatTime(item.GetUpdatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}
