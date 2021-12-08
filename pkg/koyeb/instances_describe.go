package koyeb

import (
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *InstanceHandler) Describe(cmd *cobra.Command, args []string) error {
	res, _, err := h.client.InstancesApi.GetInstance(h.ctxWithAuth, h.ResolveInstanceShortID(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	full, _ := cmd.Flags().GetBool("full")
	describeInstancesReply := NewDescribeInstanceReply(&res, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.NewMultiRenderer(
		renderer.NewDescribeRenderer(describeInstancesReply),
	).Render(output)
}

type DescribeInstanceReply struct {
	res  *koyeb.GetInstanceReply
	full bool
}

func NewDescribeInstanceReply(res *koyeb.GetInstanceReply, full bool) *DescribeInstanceReply {
	return &DescribeInstanceReply{
		res:  res,
		full: full,
	}
}

func (a *DescribeInstanceReply) MarshalBinary() ([]byte, error) {
	return a.res.GetInstance().MarshalJSON()
}

func (a *DescribeInstanceReply) Title() string {
	return "Instance"
}

func (a *DescribeInstanceReply) Headers() []string {
	return []string{"id", "status", "app", "service", "deployment_id", "datacenter", "messages"}
}

func (a *DescribeInstanceReply) Fields() []map[string]string {
	res := []map[string]string{}
	item := a.res.GetInstance()
	fields := map[string]string{
		"id":            renderer.FormatID(item.GetId(), a.full),
		"app":           renderer.FormatID(item.GetAppId(), a.full),
		"service":       renderer.FormatID(item.GetServiceId(), a.full),
		"status":        formatInstanceStatus(item.GetStatus()),
		"deployment_id": renderer.FormatID(item.GetDeploymentId(), a.full),
		"datacenter":    item.GetDatacenter(),
		"messages":      formatMessages(item.GetMessages()),
	}
	res = append(res, fields)
	return res
}

func formatMessages(msg []string) string {
	return strings.Join(msg, "\n")
}
