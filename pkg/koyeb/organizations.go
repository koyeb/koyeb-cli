package koyeb

import (
	"context"

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
		return "", errors.NewCLIErrorFromAPIError("unable to switch the current organization", err, resp)
	}
	return *res.Token.Id, nil
}
