package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *SnapshotHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string, createSnapshot *koyeb.CreateSnapshotRequest) error {
	res, resp, err := ctx.Client.SnapshotsApi.CreateSnapshot(ctx.Context).Body(*createSnapshot).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while creating the snapshot `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getSnapshotReply := NewGetSnapshotReply(ctx.Mapper, &koyeb.GetSnapshotReply{Snapshot: res.Snapshot}, full)
	ctx.Renderer.Render(getSnapshotReply)
	return nil
}
