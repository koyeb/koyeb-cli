package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *DomainHandler) Get(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	res, resp, err := ctx.Client.DomainsApi.GetDomain(ctx.Context, h.ResolveDomainArgs(ctx, args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	getDomainsReply := NewGetDomainReply(ctx.Mapper, res, full)
	ctx.Renderer.Render(getDomainsReply)
	return nil
}

type GetDomainReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetDomainReply
	full   bool
}

func NewGetDomainReply(mapper *idmapper.Mapper, value *koyeb.GetDomainReply, full bool) *GetDomainReply {
	return &GetDomainReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (GetDomainReply) Title() string {
	return "Domain"
}

func (r *GetDomainReply) MarshalBinary() ([]byte, error) {
	return r.value.GetDomain().MarshalJSON()
}

func (r *GetDomainReply) Headers() []string {
	return []string{"id", "name", "app", "status", "type", "created_at"}
}

func (r *GetDomainReply) Fields() []map[string]string {
	item := r.value.GetDomain()
	fields := map[string]string{
		"id":         renderer.FormatDomainID(r.mapper, item.GetId(), r.full),
		"name":       item.GetName(),
		"app":        renderer.FormatAppName(r.mapper, item.GetAppId(), r.full),
		"status":     string(item.GetStatus()),
		"type":       string(item.GetType()),
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}
