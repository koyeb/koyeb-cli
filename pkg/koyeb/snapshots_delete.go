package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *SnapshotHandler) Delete(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	snapshot, err := ResolveSnapshotArgs(ctx, args[0])
	if err != nil {
		return err
	}

	_, resp, err := ctx.Client.SnapshotsApi.DeleteSnapshot(ctx.Context, snapshot).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while deleting the snapshot `%s`", args[0]),
			err,
			resp,
		)
	}
	log.Infof("Snapshot %s deleted.", args[0])
	return nil
}
