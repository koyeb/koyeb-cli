package koyeb

import (
	"fmt"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) GetScale(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	serviceName, err := h.parseServiceName(cmd, args[0])
	if err != nil {
		return err
	}

	service, err := h.ResolveServiceArgs(ctx, serviceName)
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.ServicesApi.GetServiceScaling(ctx.Context, service).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving scale configuration for service `%s`", serviceName),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getScaleReply := NewGetServiceScaleReply(ctx.Mapper, res, full)
	ctx.Renderer.Render(getScaleReply)
	return nil
}

type GetServiceScaleReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetServiceScalingReply
	full   bool
}

func NewGetServiceScaleReply(mapper *idmapper.Mapper, value *koyeb.GetServiceScalingReply, full bool) *GetServiceScaleReply {
	return &GetServiceScaleReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (GetServiceScaleReply) Title() string {
	return "Service Scale Configuration"
}

func (r *GetServiceScaleReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *GetServiceScaleReply) Headers() []string {
	return []string{"instances", "scopes"}
}

func (r *GetServiceScaleReply) Fields() []map[string]string {
	scalings := r.value.GetScalings()
	resp := []map[string]string{}

	for _, scaling := range scalings {
		fields := map[string]string{
			"instances": fmt.Sprintf("%d", scaling.GetInstances()),
			"scopes":    strings.Join(scaling.GetScopes(), ", "),
		}
		resp = append(resp, fields)
	}

	return resp
}
