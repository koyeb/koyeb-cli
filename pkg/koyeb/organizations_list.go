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
	res, _, err := ctx.Client.ProfileApi.GetCurrentUser(ctx.Context).Execute()
	if err != nil {
		return "", &errors.CLIError{
			What: "The token used is not linked to a user",
			Why:  "you are authenticated with a token linked to an organization",
			Additional: []string{
				"On Koyeb, two types of tokens exist: user tokens and organization tokens.",
				"Your are currently using an organization token, which is not linked to a user.",
				"Organization tokens are unable to perform operations that require a user context, such as listing organizations or managing your account.",
			},
			Orig:     err,
			Solution: "From the Koyeb console (https://app.koyeb.com/user/settings/api/), create a user token and use it in the CLI configuration file.",
		}
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

	// ctx.Organization is empty when the field "organization" is not set in the
	// configuration file, and is not provided with the --organization flag.
	currentOrganization := ctx.Organization
	if currentOrganization == "" {
		res, resp, err := ctx.Client.ProfileApi.GetCurrentOrganization(ctx.Context).Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError("Unable to fetch the current organization", err, resp)
		}
		currentOrganization = *res.Organization.Id
	}

	full := GetBoolFlags(cmd, "full")
	reply := NewListOragnizationsReply(ctx.Mapper, &koyeb.ListOrganizationMembersReply{Members: list}, full, currentOrganization)
	ctx.Renderer.Render(reply)
	return nil
}

type ListOrganizationsReply struct {
	mapper              *idmapper.Mapper
	value               *koyeb.ListOrganizationMembersReply
	full                bool
	currentOrganization string
}

func NewListOragnizationsReply(mapper *idmapper.Mapper, value *koyeb.ListOrganizationMembersReply, full bool, currentOrganization string) *ListOrganizationsReply {
	return &ListOrganizationsReply{
		mapper:              mapper,
		value:               value,
		full:                full,
		currentOrganization: currentOrganization,
	}
}

func (ListOrganizationsReply) Title() string {
	return "Organizations"
}

func (r *ListOrganizationsReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *ListOrganizationsReply) Headers() []string {
	return []string{"id", "name", "plan", "current"}
}

func (r *ListOrganizationsReply) Fields() []map[string]string {
	items := r.value.GetMembers()
	resp := make([]map[string]string, 0, len(items))

	for _, member := range items {
		current := ""
		if member.Organization.GetId() == r.currentOrganization {
			current = "✓"
		}
		fields := map[string]string{
			"id":      renderer.FormatID(member.Organization.GetId(), r.full),
			"name":    member.Organization.GetName(),
			"plan":    string(member.Organization.GetPlan()),
			"current": current,
		}
		resp = append(resp, fields)
	}
	return resp
}
