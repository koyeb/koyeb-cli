package koyeb

import (
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func getCurrentUserId(ctx *CLIContext) (string, error) {
	res, resp, err := ctx.Client.ProfileApi.GetCurrentUser(ctx.Context).Execute()
	if err != nil {
		return "", errors.NewCLIErrorFromAPIError("The token used is not linked to a user", err, resp)
	}
	return *res.GetUser().Id, nil
}

func (h *OrganizationHandler) List(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	userId, err := getCurrentUserId(ctx)
	if err != nil {
		return err
	}

	list := []koyeb.OrganizationMember{}
	page := int64(0)
	offset := int64(0)
	limit := int64(100)

	for {
		res, resp, err := ctx.Client.OrganizationMembersApi.
			ListOrganizationMembers(ctx.Context).
			UserId(userId).
			Limit(strconv.FormatInt(limit, 10)).Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError("Error while listing organizations", err, resp)
		}
		list = append(list, res.GetMembers()...)

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	full := GetBoolFlags(cmd, "full")
	reply := NewListOragnizationsReply(ctx.Mapper, &koyeb.ListOrganizationMembersReply{Members: list}, full)
	ctx.Renderer.Render(reply)
	return nil
}

type ListOrganizationsReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.ListOrganizationMembersReply
	full   bool
}

func NewListOragnizationsReply(mapper *idmapper.Mapper, value *koyeb.ListOrganizationMembersReply, full bool) *ListOrganizationsReply {
	return &ListOrganizationsReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (ListOrganizationsReply) Title() string {
	return "Organizations"
}

func (r *ListOrganizationsReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *ListOrganizationsReply) Headers() []string {
	return []string{"id", "name", "plan"}
}

func (r *ListOrganizationsReply) Fields() []map[string]string {
	items := r.value.GetMembers()
	resp := make([]map[string]string, 0, len(items))

	for _, member := range items {
		fields := map[string]string{
			"id":   renderer.FormatID(member.Organization.GetId(), r.full),
			"name": member.Organization.GetName(),
			"plan": string(member.Organization.GetPlan()),
		}
		resp = append(resp, fields)
	}
	return resp
}
