package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) ReDeploy(cmd *cobra.Command, args []string) error {
	useCache := GetBoolFlags(cmd, "use-cache")

	redeployBody := *koyeb.NewRedeployRequestInfoWithDefaults()
	redeployBody.UseCache = &useCache
	_, resp, err := h.client.ServicesApi.ReDeploy(h.ctx, h.ResolveServiceArgs(args[0])).Info(redeployBody).Execute()

	if err != nil {
		fatalApiError(err, resp)
	}
	log.Infof("Service %s redeployed.", args[0])
	return nil
}
