package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *DomainHandler) Describe(cmd *cobra.Command, args []string) error {
	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	replies := []renderer.ApiResources{}

	getDomainRes, resp, err := h.client.DomainsApi.GetDomain(h.ctx, h.ResolveDomainArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	describeDomainsReply := NewDescribeDomainReply(h.mapper, getDomainRes, full)
	replies = append(replies, describeDomainsReply)

	// Grab attached app
	appID := getDomainRes.Domain.GetAppId()
	if appID != "" {
		getAppRes, resp, err := h.client.AppsApi.GetApp(h.ctx, appID).Execute()
		if err != nil {
			fatalApiError(err, resp)
		}

		describeAppsReply := NewDescribeAppReply(h.mapper, getAppRes, full)
		replies = append(replies, describeAppsReply)
	}

	// Grab app services if any
	if appID != "" {
		resListServices, resp, err := h.client.ServicesApi.ListServices(h.ctx).AppId(appID).Limit("100").Execute()
		if err != nil {
			fatalApiError(err, resp)
		}

		listServicesReply := NewListServicesReply(h.mapper, resListServices, full)
		replies = append(replies, listServicesReply)
	}

	return renderer.NewDescribeRenderer(replies...).Render(output)
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
		"id":          renderer.FormatDomainID(r.mapper, item.GetId(), r.full),
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
