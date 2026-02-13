package koyeb

import (
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *RegionHandler) List(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	list := []koyeb.RegionListItem{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		res, resp, err := ctx.Client.CatalogRegionsApi.ListRegions(ctx.Context).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				"Error while listing the regions",
				err,
				resp,
			)
		}
		for _, region := range res.GetRegions() {
			if strings.EqualFold(region.GetStatus(), "available") {
				list = append(list, region)
			}
		}

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	listRegionsReply := NewListRegionsReply(&koyeb.ListRegionsReply{Regions: list})
	ctx.Renderer.Render(listRegionsReply)
	return nil
}

type ListRegionsReply struct {
	value *koyeb.ListRegionsReply
}

func NewListRegionsReply(value *koyeb.ListRegionsReply) *ListRegionsReply {
	return &ListRegionsReply{
		value: value,
	}
}

func (ListRegionsReply) Title() string {
	return "Regions"
}

func (r *ListRegionsReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *ListRegionsReply) Headers() []string {
	return []string{"id", "name", "scope", "volumes_enabled"}
}

func (r *ListRegionsReply) Fields() []map[string]string {
	items := r.value.GetRegions()
	resp := make([]map[string]string, 0, len(items))

	for _, item := range items {
		fields := map[string]string{
			"id":              item.GetId(),
			"name":            item.GetName(),
			"scope":           item.GetScope(),
			"volumes_enabled": strconv.FormatBool(item.GetVolumesEnabled()),
		}
		resp = append(resp, fields)
	}

	return resp
}
