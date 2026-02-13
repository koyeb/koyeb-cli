package koyeb

import (
	"fmt"
	"os"
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

	getRegionReply := NewGetRegionReply(res)

	// For JSON/YAML output, use the standard renderer which includes instances
	// in the serialized output. For table output, print a custom format.
	if _, ok := ctx.Renderer.(*renderer.TableRenderer); ok {
		region := res.GetRegion()
		fmt.Fprintf(os.Stdout, "ID:              %s\n", region.GetId())
		fmt.Fprintf(os.Stdout, "Name:            %s\n", region.GetName())
		fmt.Fprintf(os.Stdout, "Scope:           %s\n", region.GetScope())
		fmt.Fprintf(os.Stdout, "Volumes enabled: %s\n", strconv.FormatBool(region.GetVolumesEnabled()))
		fmt.Fprintf(os.Stdout, "\nInstances:\n")
		for _, inst := range region.GetInstances() {
			fmt.Fprintf(os.Stdout, "  - %s\n", inst)
		}
	} else {
		ctx.Renderer.Render(getRegionReply)
	}
	return nil
}

type GetRegionReply struct {
	value *koyeb.GetRegionReply
}

func NewGetRegionReply(value *koyeb.GetRegionReply) *GetRegionReply {
	return &GetRegionReply{
		value: value,
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
