package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *SnapshotHandler) Get(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	snapshot, err := ResolveSnapshotArgs(ctx, args[0])
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.SnapshotsApi.GetSnapshot(ctx.Context, snapshot).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the snapshot `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getSnapshotReply := NewGetSnapshotReply(ctx.Mapper, res, full)
	ctx.Renderer.Render(getSnapshotReply)
	return nil
}

type GetSnapshotReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetSnapshotReply
	full   bool
}

func NewGetSnapshotReply(mapper *idmapper.Mapper, res *koyeb.GetSnapshotReply, full bool) *GetSnapshotReply {
	return &GetSnapshotReply{
		mapper: mapper,
		value:  res,
		full:   full,
	}
}

func (GetSnapshotReply) Title() string {
	return "Snapshot"
}

func (r *GetSnapshotReply) MarshalBinary() ([]byte, error) {
	return r.value.GetSnapshot().MarshalJSON()
}

func (r *GetSnapshotReply) Headers() []string {
	return []string{"id", "name", "region", "type", "status", "created_at", "parent_volume"}
}

func (r *GetSnapshotReply) Fields() []map[string]string {
	item := r.value.GetSnapshot()
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

	resp := []map[string]string{fields}
	return resp
}

func formatSnapshotStatus(st koyeb.SnapshotStatus) string {
	switch st {
	case koyeb.SNAPSHOTSTATUS_CREATING:
		return "creating"
	case koyeb.SNAPSHOTSTATUS_AVAILABLE:
		return "available"
	case koyeb.SNAPSHOTSTATUS_MIGRATING:
		return "migrating"
	case koyeb.SNAPSHOTSTATUS_DELETING:
		return "deleting"
	default:
		return "invalid"
	}
}

func formatSnapshotType(st koyeb.SnapshotType) string {
	switch st {
	case koyeb.SNAPSHOTTYPE_LOCAL:
		return "local"
	case koyeb.SNAPSHOTTYPE_REMOTE:
		return "remote"
	default:
		return "invalid"
	}
}
