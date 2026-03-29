package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *VolumeHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string, createVolume *koyeb.CreatePersistentVolumeRequest) error {
	projectFlag, _ := cmd.Flags().GetString("project")
	if projectFlag != "" {
		projectHandler := NewProjectHandler()
		projectID, err := projectHandler.ResolveProjectArgs(ctx, projectFlag)
		if err != nil {
			return err
		}
		createVolume.SetProjectId(projectID)
	}

	res, resp, err := ctx.Client.PersistentVolumesApi.CreatePersistentVolume(ctx.Context).Body(*createVolume).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while creating the volume `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getVolumeReply := NewGetVolumeReply(ctx.Mapper, &koyeb.GetPersistentVolumeReply{Volume: res.Volume}, full)
	ctx.Renderer.Render(getVolumeReply)
	return nil
}
