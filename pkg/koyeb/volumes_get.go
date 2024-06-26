package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *VolumeHandler) Get(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	volume, err := ResolveVolumeArgs(ctx, args[0])
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.PersistentVolumesApi.GetPersistentVolume(ctx.Context, volume).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the volume `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getVolumeReply := NewGetVolumeReply(ctx.Mapper, res, full)
	ctx.Renderer.Render(getVolumeReply)
	return nil
}

type GetVolumeReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetPersistentVolumeReply
	full   bool
}

func NewGetVolumeReply(mapper *idmapper.Mapper, res *koyeb.GetPersistentVolumeReply, full bool) *GetVolumeReply {
	return &GetVolumeReply{
		mapper: mapper,
		value:  res,
		full:   full,
	}
}

func (GetVolumeReply) Title() string {
	return "Volume"
}

func (r *GetVolumeReply) MarshalBinary() ([]byte, error) {
	return r.value.GetVolume().MarshalJSON()
}

func (r *GetVolumeReply) Headers() []string {
	return []string{"id", "name", "region", "type", "status", "size", "created_at", "service"}
}

func (r *GetVolumeReply) Fields() []map[string]string {
	item := r.value.GetVolume()
	fields := map[string]string{
		"id":         renderer.FormatID(item.GetId(), r.full),
		"name":       item.GetName(),
		"region":     item.GetRegion(),
		"status":     formatVolumeStatus(item.GetStatus()),
		"type":       formatVolumeType(item.GetBackingStore()),
		"size":       renderer.FormatSize(renderer.MBSize(item.GetMaxSize())),
		"read_only":  fmt.Sprintf("%t", item.GetReadOnly()),
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
		"updated_at": renderer.FormatTime(item.GetUpdatedAt()),
		"service":    renderer.FormatServiceSlug(r.mapper, item.GetServiceId(), r.full),
	}

	resp := []map[string]string{fields}
	return resp
}

func formatVolumeStatus(st koyeb.PersistentVolumeStatus) string {
	switch st {
	case koyeb.PERSISTENTVOLUMESTATUS_ATTACHED:
		return "attached"
	case koyeb.PERSISTENTVOLUMESTATUS_DETACHED:
		return "detached"
	case koyeb.PERSISTENTVOLUMESTATUS_DELETING:
		return "deleting"
	default:
		return "invalid"
	}
}

func formatVolumeType(st koyeb.PersistentVolumeBackingStore) string {
	switch st {
	case koyeb.PERSISTENTVOLUMEBACKINGSTORE_LOCAL_BLK:
		return "local-blk"
	default:
		return "invalid"
	}
}
