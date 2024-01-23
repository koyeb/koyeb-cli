package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *RegionalDeploymentHandler) Get(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	regionalDeployment, err := h.ResolveRegionalDeploymentArgs(ctx, args[0])
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.RegionalDeploymentsApi.GetRegionalDeployment(ctx.Context, regionalDeployment).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the regional deployment `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getRegionalDeploymentsReply := NewGetRegionalDeploymentReply(ctx.Mapper, res, full)
	ctx.Renderer.Render(getRegionalDeploymentsReply)
	return nil
}

type GetRegionalDeploymentReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetRegionalDeploymentReply
	full   bool
}

func NewGetRegionalDeploymentReply(mapper *idmapper.Mapper, value *koyeb.GetRegionalDeploymentReply, full bool) *GetRegionalDeploymentReply {
	return &GetRegionalDeploymentReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (GetRegionalDeploymentReply) Title() string {
	return "Deployment"
}

func (r *GetRegionalDeploymentReply) MarshalBinary() ([]byte, error) {
	return r.value.GetRegionalDeployment().MarshalJSON()
}

func (r *GetRegionalDeploymentReply) Headers() []string {
	return []string{"id", "app", "service", "messages", "region", "created_at"}
}

func (r *GetRegionalDeploymentReply) Fields() []map[string]string {
	item := r.value.GetRegionalDeployment()
	fields := map[string]string{
		"id":         renderer.FormatID(item.GetId(), r.full),
		"app":        renderer.FormatAppName(r.mapper, item.GetAppId(), r.full),
		"service":    renderer.FormatServiceSlug(r.mapper, item.GetServiceId(), r.full),
		"messages":   formatDeploymentMessages(item.GetMessages(), 0),
		"region":     *item.Region,
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}
