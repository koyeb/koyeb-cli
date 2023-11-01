package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Get(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	serviceName, err := parseServiceName(cmd, args[0])
	if err != nil {
		return err
	}

	service, err := h.ResolveServiceArgs(ctx, serviceName)
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.ServicesApi.GetService(ctx.Context, service).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the service `%s`", serviceName),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getServiceReply := NewGetServiceReply(ctx.Mapper, res, full)
	ctx.Renderer.Render(getServiceReply)
	return nil
}

type GetServiceReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetServiceReply
	full   bool
}

func NewGetServiceReply(mapper *idmapper.Mapper, value *koyeb.GetServiceReply, full bool) *GetServiceReply {
	return &GetServiceReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (GetServiceReply) Title() string {
	return "Service"
}

func (r *GetServiceReply) MarshalBinary() ([]byte, error) {
	return r.value.GetService().MarshalJSON()
}

func (r *GetServiceReply) Headers() []string {
	return []string{"id", "app", "name", "status", "created_at"}
}

func (r *GetServiceReply) Fields() []map[string]string {
	item := r.value.GetService()
	fields := map[string]string{
		"id":         renderer.FormatID(item.GetId(), r.full),
		"app":        renderer.FormatAppName(r.mapper, item.GetAppId(), r.full),
		"name":       item.GetName(),
		"status":     formatServiceStatus(item.GetStatus()),
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}

func formatServiceStatus(status koyeb.ServiceStatus) string {
	return string(status)
}
