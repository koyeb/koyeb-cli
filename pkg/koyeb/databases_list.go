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

func (h *DatabaseHandler) List(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	list := []koyeb.ServiceListItem{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		res, resp, err := ctx.Client.ServicesApi.ListServices(ctx.Context).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Types([]string{"DATABASE"}).
			Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				"Error while listing database services",
				err,
				resp,
			)
		}

		list = append(list, res.GetServices()...)

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	full := GetBoolFlags(cmd, "full")
	listDatabasesReply := NewListDatabasesReply(ctx.Mapper, &koyeb.ListServicesReply{Services: list}, full)
	ctx.Renderer.Render(listDatabasesReply)
	return nil
}

type ListDatabasesReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.ListServicesReply
	full   bool
}

func NewListDatabasesReply(mapper *idmapper.Mapper, value *koyeb.ListServicesReply, full bool) *ListDatabasesReply {
	return &ListDatabasesReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (ListDatabasesReply) Title() string {
	return "Services"
}

func (r *ListDatabasesReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *ListDatabasesReply) Headers() []string {
	return []string{"id", "name", "status", "created_at"}
}

func (r *ListDatabasesReply) Fields() []map[string]string {
	items := r.value.GetServices()
	resp := make([]map[string]string, 0, len(items))

	for _, item := range items {
		fields := map[string]string{
			"id":         renderer.FormatID(item.GetId(), r.full),
			"name":       item.GetName(),
			"status":     formatServiceStatus(item.GetStatus()),
			"created_at": renderer.FormatTime(item.GetCreatedAt()),
		}
		resp = append(resp, fields)
	}

	return resp
}
