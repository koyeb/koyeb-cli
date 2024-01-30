package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func (h *DatabaseHandler) Update(ctx *CLIContext, cmd *cobra.Command, args []string, serviceId string, updateService *koyeb.UpdateService) error {
	res, resp, err := ctx.Client.ServicesApi.UpdateService(ctx.Context, serviceId).Service(*updateService).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while updating the service `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getServiceReply := NewGetServiceReply(ctx.Mapper, &koyeb.GetServiceReply{Service: res.Service}, full)
	ctx.Renderer.Render(getServiceReply)
	return nil
}
