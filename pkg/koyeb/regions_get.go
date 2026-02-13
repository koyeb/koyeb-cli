package koyeb

import (
	"fmt"
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
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
	getRegionInstancesReply := NewGetRegionInstancesReply(res)
	renderer.NewChainRenderer(ctx.Renderer).
		Render(getRegionReply).
		Render(getRegionInstancesReply)
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
	return []string{"id", "name", "scope", "volumes_enabled"}
}

func (r *GetRegionReply) Fields() []map[string]string {
	item := r.value.GetRegion()
	fields := map[string]string{
		"id":              item.GetId(),
		"name":            item.GetName(),
		"scope":           item.GetScope(),
		"volumes_enabled": strconv.FormatBool(item.GetVolumesEnabled()),
	}

	return []map[string]string{fields}
}

type GetRegionInstancesReply struct {
	value *koyeb.GetRegionReply
}

func NewGetRegionInstancesReply(value *koyeb.GetRegionReply) *GetRegionInstancesReply {
	return &GetRegionInstancesReply{
		value: value,
	}
}

func (GetRegionInstancesReply) Title() string {
	return "Instances"
}

func (r *GetRegionInstancesReply) MarshalBinary() ([]byte, error) {
	return r.value.GetRegion().MarshalJSON()
}

func (r *GetRegionInstancesReply) Headers() []string {
	return []string{"instance"}
}

func (r *GetRegionInstancesReply) Fields() []map[string]string {
	region := r.value.GetRegion()
	instances := region.GetInstances()
	resp := make([]map[string]string, 0, len(instances))

	for _, inst := range instances {
		resp = append(resp, map[string]string{
			"instance": inst,
		})
	}

	return resp
}
