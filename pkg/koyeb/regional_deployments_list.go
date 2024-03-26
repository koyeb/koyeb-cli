package koyeb

import (
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *RegionalDeploymentHandler) List(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	list := []koyeb.RegionalDeploymentListItem{}

	deploymentId := ""
	if deployment, _ := cmd.Flags().GetString("deployment"); deployment != "" {
		var err error
		if deploymentId, err = h.ResolveDeploymentArgs(ctx, deployment); err != nil {
			return err
		}
	}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		req := ctx.Client.RegionalDeploymentsApi.ListRegionalDeployments(ctx.Context)

		if deploymentId != "" {
			req = req.DeploymentId(deploymentId)
		}

		res, resp, err := req.Limit(strconv.FormatInt(limit, 10)).Offset(strconv.FormatInt(offset, 10)).Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				"Error while listing regional deployments",
				err,
				resp,
			)
		}
		list = append(list, res.GetRegionalDeployments()...)

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	full := GetBoolFlags(cmd, "full")
	listRegionalDeploymentsReply := NewListRegionalDeploymentsReply(ctx.Mapper, &koyeb.ListRegionalDeploymentsReply{RegionalDeployments: list}, full)
	ctx.Renderer.Render(listRegionalDeploymentsReply)
	return nil
}

type ListRegionalDeploymentsReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.ListRegionalDeploymentsReply
	full   bool
}

func NewListRegionalDeploymentsReply(mapper *idmapper.Mapper, value *koyeb.ListRegionalDeploymentsReply, full bool) *ListRegionalDeploymentsReply {
	return &ListRegionalDeploymentsReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (ListRegionalDeploymentsReply) Title() string {
	return "Regional Deployments"
}

func (r *ListRegionalDeploymentsReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *ListRegionalDeploymentsReply) Headers() []string {
	return []string{"id", "region", "status", "messages", "created_at"}
}

func (r *ListRegionalDeploymentsReply) Fields() []map[string]string {
	items := r.value.GetRegionalDeployments()
	resp := make([]map[string]string, 0, len(items))

	for _, item := range items {
		fields := map[string]string{
			"id":         renderer.FormatID(item.GetId(), r.full),
			"region":     item.GetRegion(),
			"status":     formatRegionalDeploymentStatus(item.GetStatus()),
			"messages":   formatMessages(item.GetMessages()),
			"created_at": renderer.FormatTime(item.GetCreatedAt()),
		}
		resp = append(resp, fields)
	}

	return resp
}

func formatRegionalDeploymentStatus(status koyeb.RegionalDeploymentStatus) string {
	return string(status)
}
