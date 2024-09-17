package koyeb

import (
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *SnapshotHandler) List(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	list := []koyeb.Snapshot{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		res, resp, err := ctx.Client.SnapshotsApi.ListSnapshots(ctx.Context).
			Limit(strconv.FormatInt(limit, 10)).Offset(strconv.FormatInt(offset, 10)).Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError("Error while listing secrets", err, resp)
		}
		snapshots := res.GetSnapshots()
		if len(snapshots) == 0 {
			break
		}
		list = append(list, snapshots...)

		page++
		offset = page * limit
	}

	full := GetBoolFlags(cmd, "full")
	listSnapshotsReply := NewListSnapshotsReply(ctx.Mapper, &koyeb.ListSnapshotsReply{Snapshots: list}, full)
	ctx.Renderer.Render(listSnapshotsReply)
	return nil
}

type ListSnapshotsReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.ListSnapshotsReply
	full   bool
}

func NewListSnapshotsReply(mapper *idmapper.Mapper, value *koyeb.ListSnapshotsReply, full bool) *ListSnapshotsReply {
	return &ListSnapshotsReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (ListSnapshotsReply) Title() string {
	return "Snapshots"
}

func (r *ListSnapshotsReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *ListSnapshotsReply) Headers() []string {
	return []string{"id", "name", "region", "type", "status", "size", "created_at", "parent_volume"}
}

func (r *ListSnapshotsReply) Fields() []map[string]string {
	items := r.value.GetSnapshots()
	resp := make([]map[string]string, 0, len(items))

	for _, item := range items {
		fields := map[string]string{
			"id":            renderer.FormatID(item.GetId(), r.full),
			"name":          item.GetName(),
			"region":        item.GetRegion(),
			"status":        formatSnapshotStatus(item.GetStatus()),
			"type":          formatSnapshotType(item.GetType()),
			"created_at":    renderer.FormatTime(item.GetCreatedAt()),
			"updated_at":    renderer.FormatTime(item.GetUpdatedAt()),
			"parent_volume": renderer.FormatID(item.GetParentVolumeId(), r.full),
		}
		resp = append(resp, fields)
	}

	return resp
}
