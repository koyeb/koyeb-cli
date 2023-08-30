package koyeb

import (
	"context"
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func NewOrganizationCmd() *cobra.Command {
	h := NewOrganizationHandler()
	rootCmd := &cobra.Command{
		Use:     "organizations ACTION",
		Aliases: []string{"organizations", "organization", "orgas", "orga", "orgs", "org", "organisations", "organisation"},
		Short:   "Organization",
	}
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List organizations",
		RunE:  WithCLIContext(h.List),
	}
	rootCmd.AddCommand(listCmd)

	switchCmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch the CLI context to another organization",
		RunE:  WithCLIContext(h.Switch),
	}
	rootCmd.AddCommand(switchCmd)
	return rootCmd
}

func NewOrganizationHandler() *OrganizationHandler {
	return &OrganizationHandler{}
}

type OrganizationHandler struct {
}

func ResolveOrganizationArgs(ctx *CLIContext, val string) (string, error) {
	organizationMapper := ctx.Mapper.Organization()
	id, err := organizationMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}

// GetOrganizationToken calls /v1/organizations/{organizationId}/switch which returns a token to access the resources of organizationId
func GetOrganizationToken(api koyeb.OrganizationApi, ctx context.Context, organizationId string) (string, error) {
	//SwitchOrganization requires to pass an empty body
	body := make(map[string]interface{})
	res, resp, err := api.SwitchOrganization(ctx, organizationId).Body(body).Execute()
	if err != nil {
		errBuf := make([]byte, 1024)
		//  if the body can't be read, it won't be displayed in the error below
		resp.Body.Read(errBuf) // nolint:errcheck
		return "", &errors.CLIError{
			What: "Error while switching the current organization",
			Why:  fmt.Sprintf("the API endpoint which switches the current organization returned an error %d", resp.StatusCode),
			Additional: []string{
				"You provided an organization id with the --organization flag, or the `organization` field is set in your configuration file.",
				"The value provided is likely incorrect, or you don't have access to this organization.",
			},
			Orig:     fmt.Errorf("HTTP/%d\n\n%s", resp.StatusCode, errBuf),
			Solution: "List your organizations with `koyeb --organization=\"\" organization list`, then switch to the organization you want to use with `koyeb --organization=\"\" organization switch <id>`. Finally you can run your commands again, without the --organization flag.",
		}
	}
	return *res.Token.Id, nil
}
