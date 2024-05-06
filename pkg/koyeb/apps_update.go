package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Update(ctx *CLIContext, cmd *cobra.Command, args []string, updateApp *koyeb.UpdateApp) error {
	app, err := h.ResolveAppArgs(ctx, args[0])
	if err != nil {
		return err
	}

	appRes, appResp, err := ctx.Client.AppsApi.UpdateApp2(ctx.Context, app).App(*updateApp).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while updating the application `%s`", args[0]),
			err,
			appResp,
		)
	}

	var newDomain *koyeb.Domain
	if domain := GetStringFlags(cmd, "domain"); domain != "" {
		updateDomainReq := koyeb.NewUpdateDomainWithDefaults()
		updateDomainReq.SetAppId(app)
		updateDomainReq.SetSubdomain(domain)

		domainID, err := ctx.Mapper.App().GetAutoDomain(app)
		if err != nil {
			return &errors.CLIError{
				What:       fmt.Sprintf("Error while renaming the automatic domain for `%s` to %q", args[0], domain),
				Why:        "Could not find the automatic domain for the application",
				Additional: []string{"This could be a temporary error"},
				Orig:       err,
				Solution:   errors.SolutionTryAgainOrUpdateOrIssue,
			}
		}

		domainRes, domainResp, err := ctx.Client.DomainsApi.UpdateDomain(ctx.Context, domainID).Domain(*updateDomainReq).Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				fmt.Sprintf("Error while renaming the automatic domain for `%s` to %q", args[0], args[1]),
				err,
				domainResp,
			)
		}

		if domainRes.Domain != nil {
			newDomain = domainRes.Domain
		}
	}

	full := GetBoolFlags(cmd, "full")
	if newDomain != nil {
		appRes.App.Domains = append(appRes.App.Domains, *newDomain)
	}
	getAppsReply := NewGetAppReply(ctx.Mapper, &koyeb.GetAppReply{App: appRes.App}, full)
	ctx.Renderer.Render(getAppsReply)
	return nil
}
