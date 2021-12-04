package koyeb

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Update(cmd *cobra.Command, args []string, updateService *koyeb.UpdateService) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	res, _, err := client.ServicesApi.UpdateService(ctx, ResolveAppShortID(args[0])).Body(*updateService).Execute()
	if err != nil {
		fatalApiError(err)
	}
	log.Infof("Service deployment in progress. Access deployment logs running: koyeb service logs %s.", res.Service.GetId()[:8])
	full, _ := cmd.Flags().GetBool("full")
	getServiceReply := NewGetServiceReply(&koyeb.GetServiceReply{Service: res.Service}, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.DescribeRenderer(output, getServiceReply)
}
