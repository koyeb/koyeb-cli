package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *SnapshotHandler) Update(ctx *CLIContext, cmd *cobra.Command, args []string, snapshot *koyeb.UpdateSnapshotRequest) error {
	id, err := ResolveSnapshotArgs(ctx, args[0])
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.SnapshotsApi.UpdateSnapshot(ctx.Context, id).Body(*snapshot).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while updating the snapshot `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getSnapshotReply := NewGetSnapshotReply(ctx.Mapper, &koyeb.GetSnapshotReply{Snapshot: res.Snapshot}, full)
	ctx.Renderer.Render(getSnapshotReply)
	return nil
}
