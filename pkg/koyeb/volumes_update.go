package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *VolumeHandler) Update(ctx *CLIContext, cmd *cobra.Command, args []string, volume *koyeb.UpdatePersistentVolumeRequest) error {
	id, err := ResolveVolumeArgs(ctx, args[0])
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.PersistentVolumesApi.UpdatePersistentVolume(ctx.Context, id).Body(*volume).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while updating the volume `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getVolumeReply := NewGetVolumeReply(ctx.Mapper, &koyeb.GetPersistentVolumeReply{Volume: res.Volume}, full)
	ctx.Renderer.Render(getVolumeReply)
	return nil
}
