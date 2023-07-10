package koyeb

import (
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *InstanceHandler) List(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	query, err := h.getListQuery(ctx, cmd)
	if err != nil {
		return err
	}

	list := []koyeb.InstanceListItem{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		res, resp, err := query.Limit(strconv.FormatInt(limit, 10)).Offset(strconv.FormatInt(offset, 10)).Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				"Error while listing the instances",
				err,
				resp,
			)
		}
		list = append(list, res.GetInstances()...)

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	full := GetBoolFlags(cmd, "full")
	listInstancesReply := NewListInstancesReply(ctx.Mapper, &koyeb.ListInstancesReply{Instances: list}, full)
	ctx.Renderer.Render(listInstancesReply)
	return nil
}

func (h *InstanceHandler) getListQuery(ctx *CLIContext, cmd *cobra.Command) (koyeb.ApiListInstancesRequest, error) {
	query := ctx.Client.InstancesApi.ListInstances(ctx.Context).Statuses([]string{
		string(koyeb.INSTANCESTATUS_ALLOCATING),
		string(koyeb.INSTANCESTATUS_STARTING),
		string(koyeb.INSTANCESTATUS_HEALTHY),
		string(koyeb.INSTANCESTATUS_UNHEALTHY),
		string(koyeb.INSTANCESTATUS_STOPPING),
	})

	query, err := h.getAppIDForListQuery(ctx, query, GetStringFlags(cmd, "app"))
	if err != nil {
		return query, err
	}

	query, err = h.getServiceIDForListQuery(ctx, query, GetStringFlags(cmd, "service"))
	if err != nil {
		return query, err
	}

	return query, nil
}

func (h *InstanceHandler) getAppIDForListQuery(ctx *CLIContext, query koyeb.ApiListInstancesRequest, filter string) (koyeb.ApiListInstancesRequest, error) {
	if filter == "" {
		return query, nil
	}

	id, err := ctx.Mapper.App().ResolveID(filter)
	if err != nil {
		return query, err
	}
	return query.AppId(id), nil
}

func (h *InstanceHandler) getServiceIDForListQuery(ctx *CLIContext, query koyeb.ApiListInstancesRequest, filter string) (koyeb.ApiListInstancesRequest, error) {
	if filter == "" {
		return query, nil
	}

	id, err := ctx.Mapper.Service().ResolveID(filter)
	if err != nil {
		return query, err
	}
	return query.ServiceId(id), nil
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
			"id":         renderer.FormatID(item.GetId(), r.full),
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
