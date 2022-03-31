package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
)

type ListRegionalDeploymentsReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.ListRegionalDeploymentsReply
	full   bool
}

func NewListRegionalDeploymentsReply(mapper *idmapper.Mapper, value *koyeb.ListRegionalDeploymentsReply, full bool) *ListRegionalDeploymentsReply {
	return &ListRegionalDeploymentsReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (ListRegionalDeploymentsReply) Title() string {
	return "Regional Deployments"
}

func (r *ListRegionalDeploymentsReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *ListRegionalDeploymentsReply) Headers() []string {
	return []string{"id", "region", "status", "messages", "created_at"}
}

func (r *ListRegionalDeploymentsReply) Fields() []map[string]string {
	items := r.value.GetRegionalDeployments()
	resp := make([]map[string]string, 0, len(items))

	for _, item := range items {
		fields := map[string]string{
			"id":         renderer.FormatRegionalDeploymentID(r.mapper, item.GetId(), r.full),
			"region":     item.GetRegion(),
			"status":     formatRegionalDeploymentStatus(item.GetStatus()),
			"messages":   formatMessages(item.GetMessages()),
			"created_at": renderer.FormatTime(item.GetCreatedAt()),
		}
		resp = append(resp, fields)
	}

	return resp
}

func formatRegionalDeploymentStatus(status koyeb.RegionalDeploymentStatus) string {
	return string(status)
}
