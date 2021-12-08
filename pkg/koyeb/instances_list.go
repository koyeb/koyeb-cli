package koyeb

import (
	"context"
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *InstanceHandler) getListQuery(ctx context.Context, cmd *cobra.Command, client *koyeb.APIClient, appMapper *idmapper.AppMapper, serviceMapper *idmapper.ServiceMapper) koyeb.ApiListInstancesRequest {
	appFilter, _ := cmd.Flags().GetString("app")
	serviceFilter, _ := cmd.Flags().GetString("service")
	appID := ""

	query := client.InstancesApi.ListInstances(ctx).Statuses([]string{
		string(koyeb.INSTANCESTATUS_ALLOCATING),
		string(koyeb.INSTANCESTATUS_STARTING),
		string(koyeb.INSTANCESTATUS_HEALTHY),
		string(koyeb.INSTANCESTATUS_UNHEALTHY),
		string(koyeb.INSTANCESTATUS_STOPPING),
	})

	query, appID = h.getAppIDForListQuery(query, appFilter, appMapper)
	query = h.getServiceIDForListQuery(query, appID, serviceFilter, serviceMapper)

	return query
}

func (h *InstanceHandler) List(cmd *cobra.Command, args []string) error {
	ctx := h.ctxWithAuth

	appMapper := idmapper.NewAppMapper(ctx, h.client)
	serviceMapper := idmapper.NewServiceMapper(ctx, h.client)

	query := h.getListQuery(ctx, cmd, h.client, appMapper, serviceMapper)
	results := koyeb.ListInstancesReply{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		resp, _, err := query.Limit(strconv.FormatInt(limit, 10)).Offset(strconv.FormatInt(offset, 10)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		if results.Instances == nil {
			results = resp
		} else {
			*results.Instances = append(*results.Instances, *resp.Instances...)
		}
		page += 1
		offset = page * limit
		if offset >= resp.GetCount() {
			break
		}
	}
	full, _ := cmd.Flags().GetBool("full")
	listInstancesReply := NewListInstancesReply(results, appMapper, serviceMapper, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.NewListRenderer(listInstancesReply).Render(output)
}

type ListInstancesReply struct {
	items         koyeb.ListInstancesReply
	appMapper     *idmapper.AppMapper
	serviceMapper *idmapper.ServiceMapper
	full          bool
}

func NewListInstancesReply(items koyeb.ListInstancesReply, appMapper *idmapper.AppMapper, serviceMapper *idmapper.ServiceMapper, full bool) *ListInstancesReply {
	return &ListInstancesReply{
		full:          full,
		items:         items,
		appMapper:     appMapper,
		serviceMapper: serviceMapper,
	}
}

func (ListInstancesReply) Title() string {
	return "Instances"
}

func (reply *ListInstancesReply) MarshalBinary() ([]byte, error) {
	return reply.items.MarshalJSON()
}

func (reply *ListInstancesReply) Headers() []string {
	return []string{"id", "status", "app", "service", "deployment_id", "datacenter"}
}

func (reply *ListInstancesReply) Fields() []map[string]string {
	items := reply.items.GetInstances()
	resp := make([]map[string]string, 0, len(items))

	for _, item := range items {

		fields := map[string]string{
			"id":            renderer.FormatID(item.GetId(), reply.full),
			"app":           renderer.FormatID(item.GetAppId(), reply.full),
			"service":       renderer.FormatID(item.GetServiceId(), reply.full),
			"status":        formatInstanceStatus(item.GetStatus()),
			"deployment_id": renderer.FormatID(item.GetDeploymentId(), reply.full),
			"datacenter":    item.GetDatacenter(),
		}
		resp = append(resp, fields)
	}

	return resp
}
