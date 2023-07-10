package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *DomainHandler) Describe(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	domain, err := h.ResolveDomainArgs(ctx, args[0])
	if err != nil {
		return err
	}

	full := GetBoolFlags(cmd, "full")
	replies := []renderer.ApiResources{}

	getDomainRes, resp, err := ctx.Client.DomainsApi.GetDomain(ctx.Context, domain).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the domain `%s`", args[0]),
			err,
			resp,
		)
	}

	describeDomainsReply := NewDescribeDomainReply(ctx.Mapper, getDomainRes, full)
	replies = append(replies, describeDomainsReply)

	// Grab attached app
	appID := getDomainRes.Domain.GetAppId()
	if appID != "" {
		getAppRes, resp, err := ctx.Client.AppsApi.GetApp(ctx.Context, appID).Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				fmt.Sprintf("Error while retrieving the application `%s`", appID),
				err,
				resp,
			)
		}

		describeAppsReply := NewDescribeAppReply(ctx.Mapper, getAppRes, full)
		replies = append(replies, describeAppsReply)
	}

	// Grab app services if any
	if appID != "" {
		resListServices, resp, err := ctx.Client.ServicesApi.ListServices(ctx.Context).AppId(appID).Limit("100").Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				fmt.Sprintf("Error while retrieving the services for the application `%s`", appID),
				err,
				resp,
			)
		}

		listServicesReply := NewListServicesReply(ctx.Mapper, resListServices, full)
		replies = append(replies, listServicesReply)
	}

	renderer := renderer.NewChainRenderer(ctx.Renderer)
	for _, reply := range replies {
		renderer.Render(reply)
	}
	return nil
}

type DescribeDomainReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetDomainReply
	full   bool
}

func NewDescribeDomainReply(mapper *idmapper.Mapper, value *koyeb.GetDomainReply, full bool) *DescribeDomainReply {
	return &DescribeDomainReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (DescribeDomainReply) Title() string {
	return "Domain"
}

func (r *DescribeDomainReply) MarshalBinary() ([]byte, error) {
	return r.value.GetDomain().MarshalJSON()
}

func (r *DescribeDomainReply) Headers() []string {
	return []string{"id", "name", "app", "status", "type", "created_at", "updated_at", "verified_at", "messages"}
}

func (r *DescribeDomainReply) Fields() []map[string]string {
	item := r.value.GetDomain()

	fields := map[string]string{
		"id":          renderer.FormatID(item.GetId(), r.full),
		"name":        item.GetName(),
		"app":         renderer.FormatAppName(r.mapper, item.GetAppId(), r.full),
		"status":      string(item.GetStatus()),
		"type":        string(item.GetType()),
		"created_at":  renderer.FormatTime(item.GetCreatedAt()),
		"updated_at":  renderer.FormatTime(item.GetUpdatedAt()),
		"verified_at": formatVerifiedAt(&item),
		"messages":    formatMessages(item.GetMessages()),
	}

	resp := []map[string]string{fields}
	return resp
}

func formatVerifiedAt(domain *koyeb.Domain) string {
	verifiedAt, ok := domain.GetVerifiedAtOk()
	if !ok {
		return ""
	}

	return renderer.FormatTime(*verifiedAt)
}
