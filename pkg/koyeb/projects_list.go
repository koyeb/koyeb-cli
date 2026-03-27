package koyeb

import (
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *ProjectHandler) List(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	list := []koyeb.Project{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		res, resp, err := ctx.Client.ProjectsApi.ListProjects(ctx.Context).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError("Error while listing projects", err, resp)
		}

		projects := res.GetProjects()
		if len(projects) == 0 {
			break
		}

		list = append(list, projects...)
		page++
		offset = page * limit
		if int64(len(projects)) < limit {
			break
		}
	}

	full := GetBoolFlags(cmd, "full")
	ctx.Renderer.Render(NewListProjectsReply(&koyeb.ListProjectsReply{Projects: list}, full))
	return nil
}

type ListProjectsReply struct {
	value *koyeb.ListProjectsReply
	full  bool
}

func NewListProjectsReply(value *koyeb.ListProjectsReply, full bool) *ListProjectsReply {
	return &ListProjectsReply{
		value: value,
		full:  full,
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
		resp = append(resp, projectFields(item, r.full))
	}

	return resp
}
