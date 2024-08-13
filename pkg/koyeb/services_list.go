package koyeb

import (
	"fmt"
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) List(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	list := []koyeb.ServiceListItem{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		req := ctx.Client.ServicesApi.ListServices(ctx.Context)
		appId := GetStringFlags(cmd, "app")

		if appId != "" {
			app, err := h.ResolveAppArgs(ctx, appId)
			if err != nil {
				return err
			}
			req = req.AppId(app)
		}
		name := GetStringFlags(cmd, "name")
		if name != "" {
			req = req.Name(name)
		}
		res, resp, err := req.Limit(strconv.FormatInt(limit, 10)).Offset(strconv.FormatInt(offset, 10)).Execute()
		if err != nil {
			errTitle := "Error while listing services"
			if appId != "" {
				errTitle = fmt.Sprintf("Error while listing the services of the application `%s`", appId)
			}
			return errors.NewCLIErrorFromAPIError(errTitle, err, resp)
		}

		for _, svc := range res.GetServices() {
			// Database services are displayed in the database command
			if svc.GetType() == koyeb.SERVICETYPE_DATABASE {
				continue
			}
			list = append(list, svc)
		}

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	full := GetBoolFlags(cmd, "full")
	listServicesReply := NewListServicesReply(ctx.Mapper, &koyeb.ListServicesReply{Services: list}, full)
	ctx.Renderer.Render(listServicesReply)
	return nil
}

type ListServicesReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.ListServicesReply
	full   bool
}

func NewListServicesReply(mapper *idmapper.Mapper, value *koyeb.ListServicesReply, full bool) *ListServicesReply {
	return &ListServicesReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (ListServicesReply) Title() string {
	return "Services"
}

func (r *ListServicesReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *ListServicesReply) Headers() []string {
	return []string{"id", "app", "name", "type", "status", "created_at"}
}

func (r *ListServicesReply) Fields() []map[string]string {
	items := r.value.GetServices()
	resp := make([]map[string]string, 0, len(items))

	for _, item := range items {
		fields := map[string]string{
			"id":         renderer.FormatID(item.GetId(), r.full),
			"app":        renderer.FormatAppName(r.mapper, item.GetAppId(), r.full),
			"name":       item.GetName(),
			"type":       string(item.GetType()),
			"status":     formatServiceStatus(item.GetStatus()),
			"created_at": renderer.FormatTime(item.GetCreatedAt()),
		}
		resp = append(resp, fields)
	}

	return resp
}
