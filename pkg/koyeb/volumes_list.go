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

func (h *VolumeHandler) List(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	list := []koyeb.PersistentVolume{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		res, resp, err := ctx.Client.PersistentVolumesApi.ListPersistentVolumes(ctx.Context).
			Limit(strconv.FormatInt(limit, 10)).Offset(strconv.FormatInt(offset, 10)).Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError("Error while listing secrets", err, resp)
		}
		volumes := res.GetVolumes()
		if len(volumes) == 0 {
			break
		}
		list = append(list, volumes...)

		page++
		offset = page * limit
	}

	full := GetBoolFlags(cmd, "full")
	listVolumesReply := NewListVolumesReply(ctx.Mapper, &koyeb.ListPersistentVolumesReply{Volumes: list}, full)
	ctx.Renderer.Render(listVolumesReply)
	return nil
}

type ListVolumesReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.ListPersistentVolumesReply
	full   bool
}

func NewListVolumesReply(mapper *idmapper.Mapper, value *koyeb.ListPersistentVolumesReply, full bool) *ListVolumesReply {
	return &ListVolumesReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (ListVolumesReply) Title() string {
	return "Volumes"
}

func (r *ListVolumesReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *ListVolumesReply) Headers() []string {
	return []string{"id", "name", "region", "type", "status", "size", "created_at", "service"}
}

func (r *ListVolumesReply) Fields() []map[string]string {
	items := r.value.GetVolumes()
	resp := make([]map[string]string, 0, len(items))

	for _, item := range items {
		fields := map[string]string{
			"id":         renderer.FormatID(item.GetId(), r.full),
			"name":       item.GetName(),
			"region":     item.GetRegion(),
			"status":     formatVolumeStatus(item.GetStatus()),
			"type":       formatVolumeType(item.GetBackingStore()),
			"size":       renderer.FormatSize(renderer.GBSize(item.GetMaxSize())),
			"read_only":  fmt.Sprintf("%t", item.GetReadOnly()),
			"created_at": renderer.FormatTime(item.GetCreatedAt()),
			"updated_at": renderer.FormatTime(item.GetUpdatedAt()),
			"service":    renderer.FormatServiceSlug(r.mapper, item.GetServiceId(), r.full),
		}
		resp = append(resp, fields)
	}

	return resp
}
