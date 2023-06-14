package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) ReDeploy(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	useCache := GetBoolFlags(cmd, "use-cache")

	redeployBody := *koyeb.NewRedeployRequestInfoWithDefaults()
	redeployBody.UseCache = &useCache
	_, resp, err := ctx.client.ServicesApi.ReDeploy(ctx.context, h.ResolveServiceArgs(ctx, args[0])).Info(redeployBody).Execute()

	if err != nil {
		fatalApiError(err, resp)
	}
	log.Infof("Service %s redeployed.", args[0])
	return nil
}
