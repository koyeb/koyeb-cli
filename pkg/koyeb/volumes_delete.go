package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *VolumeHandler) Delete(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	volume, err := ResolveVolumeArgs(ctx, args[0])
	if err != nil {
		return err
	}

	_, resp, err := ctx.Client.PersistentVolumesApi.DeletePersistentVolume(ctx.Context, volume).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while deleting the volume `%s`", args[0]),
			err,
			resp,
		)
	}
	log.Infof("Volume %s deleted.", args[0])
	return nil
}
