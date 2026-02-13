package koyeb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *RegionHandler) Get(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	res, resp, err := ctx.Client.CatalogRegionsApi.GetRegion(ctx.Context, args[0]).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the region `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getRegionReply := NewGetRegionReply(res, full)
	ctx.Renderer.Render(getRegionReply)
	return nil
}

type GetRegionReply struct {
	value *koyeb.GetRegionReply
	full  bool
}

func NewGetRegionReply(value *koyeb.GetRegionReply, full bool) *GetRegionReply {
	return &GetRegionReply{
		value: value,
		full:  full,
	}
}

func (GetRegionReply) Title() string {
	return "Region"
}

func (r *GetRegionReply) MarshalBinary() ([]byte, error) {
	return r.value.GetRegion().MarshalJSON()
}

func (r *GetRegionReply) Headers() []string {
	return []string{"id", "name", "scope", "volumes_enabled", "instances"}
}

func (r *GetRegionReply) Fields() []map[string]string {
	item := r.value.GetRegion()
	fields := map[string]string{
		"id":              item.GetId(),
		"name":            item.GetName(),
		"scope":           item.GetScope(),
		"volumes_enabled": strconv.FormatBool(item.GetVolumesEnabled()),
		"instances":       strings.Join(item.GetInstances(), ", "),
	}

	return []map[string]string{fields}
}
