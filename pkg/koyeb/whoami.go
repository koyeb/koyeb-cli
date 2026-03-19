package koyeb

import (
	"encoding/json"
	"net/http"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewWhoAmICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show information about the currently authenticated user or organization",
		RunE:  WithCLIContext(WhoAmI),
	}
}

func WhoAmI(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	full := GetBoolFlags(cmd, "full")

	// Try to get the current user. This only succeeds with a user token.
	userRes, userResp, userErr := ctx.Client.ProfileApi.GetCurrentUser(ctx.Context).Execute()
	isUserToken := userErr == nil
	log.Debugf("whoami: GetCurrentUser: isUserToken=%v status=%s", isUserToken, httpStatus(userResp))
	log.Debugf("whoami: ctx.Organization=%q", ctx.Organization)

	var org *koyeb.Organization

	if ctx.Organization != "" {
		// Org ID is known (set via --organization flag or config).
		// GetOrganization works for any valid token and returns 401 on invalid ones.
		log.Debugf("whoami: fetching org via GetOrganization(%s)", ctx.Organization)
		orgRes, resp, err := ctx.Client.OrganizationApi.GetOrganization(ctx.Context, ctx.Organization).Execute()
		if err != nil {
			log.Debugf("whoami: GetOrganization failed: status=%s err=%v", httpStatus(resp), err)
			return errors.NewCLIErrorFromAPIError("Error while retrieving the current organization", err, resp)
		}
		o := orgRes.GetOrganization()
		org = &o
	} else if isUserToken {
		// User token: the profile endpoint returns the current org and errors on invalid tokens.
		log.Debugf("whoami: fetching org via GetCurrentOrganization")
		orgRes, resp, err := ctx.Client.ProfileApi.GetCurrentOrganization(ctx.Context).Execute()
		if err != nil {
			log.Debugf("whoami: GetCurrentOrganization failed: status=%s err=%v", httpStatus(resp), err)
			return errors.NewCLIErrorFromAPIError("Error while retrieving the current organization", err, resp)
		}
		o := orgRes.GetOrganization()
		org = &o
	} else {
		// Direct org API key with no org ID in context.
		// Discover the org ID from apps (which always carry organization_id).
		// A 401 here means the token is invalid or missing.
		log.Debugf("whoami: org token with no org ID in context — discovering via ListApps")
		appsRes, resp, err := ctx.Client.AppsApi.ListApps(ctx.Context).Limit("1").Execute()
		if err != nil {
			log.Debugf("whoami: ListApps failed: status=%s err=%v", httpStatus(resp), err)
			return errors.NewCLIErrorFromAPIError("Error while validating the token", err, resp)
		}
		apps := appsRes.GetApps()
		if len(apps) > 0 {
			if orgId := apps[0].GetOrganizationId(); orgId != "" {
				log.Debugf("whoami: discovered org ID %q from apps", orgId)
				orgRes, resp, err := ctx.Client.OrganizationApi.GetOrganization(ctx.Context, orgId).Execute()
				if err != nil {
					log.Debugf("whoami: GetOrganization(%s) failed: status=%s err=%v", orgId, httpStatus(resp), err)
				} else {
					o := orgRes.GetOrganization()
					org = &o
				}
			}
		} else {
			log.Debugf("whoami: no apps found, org fields will be empty")
		}
	}

	var user *koyeb.User
	if isUserToken {
		u := userRes.GetUser()
		user = &u
	}

	ctx.Renderer.Render(NewWhoAmIReply(user, org, full))
	return nil
}

func httpStatus(resp *http.Response) string {
	if resp == nil {
		return "no response"
	}
	return http.StatusText(resp.StatusCode)
}

type WhoAmIReply struct {
	user *koyeb.User
	org  *koyeb.Organization
	full bool
}

func NewWhoAmIReply(user *koyeb.User, org *koyeb.Organization, full bool) *WhoAmIReply {
	return &WhoAmIReply{user: user, org: org, full: full}
}

func (WhoAmIReply) Title() string {
	return "Identity"
}

func (r *WhoAmIReply) MarshalBinary() ([]byte, error) {
	tokenType := "organization"
	if r.user != nil {
		tokenType = "user"
	}
	out := map[string]interface{}{
		"token_type": tokenType,
		"org_id":     "",
		"org_name":   "",
		"plan":       "",
		"org_status": "",
	}
	if r.user != nil {
		out["user_id"] = r.user.GetId()
		out["name"] = r.user.GetName()
		out["email"] = r.user.GetEmail()
	}
	if r.org != nil {
		out["org_id"] = r.org.GetId()
		out["org_name"] = r.org.GetName()
		out["plan"] = string(r.org.GetPlan())
		out["org_status"] = string(r.org.GetStatus())
	}
	return json.Marshal(out)
}

func (r *WhoAmIReply) Headers() []string {
	if r.user != nil {
		return []string{"token_type", "org_id", "org_name", "plan", "org_status", "user_id", "name", "email"}
	}
	return []string{"token_type", "org_id", "org_name", "plan", "org_status"}
}

func (r *WhoAmIReply) Fields() []map[string]string {
	tokenType := "organization"
	if r.user != nil {
		tokenType = "user"
	}
	fields := map[string]string{
		"token_type": tokenType,
		"org_id":     "",
		"org_name":   "",
		"plan":       "",
		"org_status": "",
	}
	if r.user != nil {
		fields["user_id"] = renderer.FormatID(r.user.GetId(), r.full)
		fields["name"] = r.user.GetName()
		fields["email"] = r.user.GetEmail()
	}
	if r.org != nil {
		fields["org_id"] = renderer.FormatID(r.org.GetId(), r.full)
		fields["org_name"] = r.org.GetName()
		fields["plan"] = string(r.org.GetPlan())
		fields["org_status"] = string(r.org.GetStatus())
	}
	return []map[string]string{fields}
}
