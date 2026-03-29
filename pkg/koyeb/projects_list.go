package koyeb

import (
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *ProjectHandler) List(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	list := []koyeb.Project{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		res, resp, err := ctx.Client.ProjectsApi.ListProjects(ctx.Context).
			Limit(strconv.FormatInt(limit, 10)).Offset(strconv.FormatInt(offset, 10)).Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				"Error while listing the projects",
				err,
				resp,
			)
		}
		list = append(list, res.GetProjects()...)

		// If we got fewer items than requested, we've reached the end
		if int64(len(res.GetProjects())) < limit {
			break
		}

		page++
		offset = page * limit
	}

	full := GetBoolFlags(cmd, "full")
	listProjectsReply := NewListProjectsReply(ctx.Mapper, &koyeb.ListProjectsReply{Projects: list}, full)
	ctx.Renderer.Render(listProjectsReply)
	return nil
}

type ListProjectsReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.ListProjectsReply
	full   bool
}

func NewListProjectsReply(mapper *idmapper.Mapper, value *koyeb.ListProjectsReply, full bool) *ListProjectsReply {
	return &ListProjectsReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (ListProjectsReply) Title() string {
	return "Projects"
}

func (r *ListProjectsReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *ListProjectsReply) Headers() []string {
	return []string{"id", "name", "description", "created_at"}
}

func (r *ListProjectsReply) Fields() []map[string]string {
	items := r.value.GetProjects()
	resp := make([]map[string]string, 0, len(items))

	for _, item := range items {
		fields := map[string]string{
			"id":          renderer.FormatID(item.GetId(), r.full),
			"name":        item.GetName(),
			"description": item.GetDescription(),
			"created_at":  renderer.FormatTime(item.GetCreatedAt()),
		}
		resp = append(resp, fields)
	}

	return resp
}
