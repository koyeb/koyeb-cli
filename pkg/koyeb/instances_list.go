package koyeb

import (
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *InstanceHandler) List(cmd *cobra.Command, args []string) error {
	query := h.getListQuery(cmd)
	list := []koyeb.InstanceListItem{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		res, resp, err := query.Limit(strconv.FormatInt(limit, 10)).Offset(strconv.FormatInt(offset, 10)).Execute()
		if err != nil {
			fatalApiError(err, resp)
		}
		list = append(list, res.GetInstances()...)

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	listInstancesReply := NewListInstancesReply(h.mapper, &koyeb.ListInstancesReply{Instances: list}, full)

	return renderer.NewListRenderer(listInstancesReply).Render(output)
}

func (h *InstanceHandler) getListQuery(cmd *cobra.Command) koyeb.ApiListInstancesRequest {
	query := h.client.InstancesApi.ListInstances(h.ctx).Statuses([]string{
		string(koyeb.INSTANCESTATUS_ALLOCATING),
		string(koyeb.INSTANCESTATUS_STARTING),
		string(koyeb.INSTANCESTATUS_HEALTHY),
		string(koyeb.INSTANCESTATUS_UNHEALTHY),
		string(koyeb.INSTANCESTATUS_STOPPING),
	})

	query = h.getAppIDForListQuery(query, GetStringFlags(cmd, "app"))
	query = h.getServiceIDForListQuery(query, GetStringFlags(cmd, "service"))

	return query
}

func (h *InstanceHandler) getAppIDForListQuery(query koyeb.ApiListInstancesRequest, filter string) koyeb.ApiListInstancesRequest {
	if filter == "" {
		return query
	}

	id, err := h.mapper.App().ResolveID(filter)
	if err != nil {
		fatalApiError(err, nil)
	}

	return query.AppId(id)
}

func (h *InstanceHandler) getServiceIDForListQuery(query koyeb.ApiListInstancesRequest, filter string) koyeb.ApiListInstancesRequest {
	if filter == "" {
		return query
	}

	id, err := h.mapper.Service().ResolveID(filter)
	if err != nil {
		fatalApiError(err, nil)
	}

	return query.ServiceId(id)
}

type ListInstancesReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.ListInstancesReply
	full   bool
}

func NewListInstancesReply(mapper *idmapper.Mapper, value *koyeb.ListInstancesReply, full bool) *ListInstancesReply {
	return &ListInstancesReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (ListInstancesReply) Title() string {
	return "Instances"
}

func (r *ListInstancesReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *ListInstancesReply) Headers() []string {
	return []string{"id", "service", "status", "region", "datacenter", "created_at"}
}

func (r *ListInstancesReply) Fields() []map[string]string {
	items := r.value.GetInstances()
	resp := make([]map[string]string, 0, len(items))

	for _, item := range items {
		fields := map[string]string{
			"id":         renderer.FormatInstanceID(r.mapper, item.GetId(), r.full),
			"service":    renderer.FormatServiceSlug(r.mapper, item.GetServiceId(), r.full),
			"status":     formatInstanceStatus(item.GetStatus()),
			"region":     item.GetRegion(),
			"datacenter": item.GetDatacenter(),
			"created_at": renderer.FormatTime(item.GetCreatedAt()),
		}
		resp = append(resp, fields)
	}

	return resp
}
