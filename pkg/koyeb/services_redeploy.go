package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) ReDeploy(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	serviceName, err := parseServiceName(cmd, args[0])
	if err != nil {
		return err
	}

	service, err := h.ResolveServiceArgs(ctx, serviceName)
	if err != nil {
		return err
	}

	useCache := GetBoolFlags(cmd, "use-cache")
	skipBuild := GetBoolFlags(cmd, "skip-build")

	redeployBody := *koyeb.NewRedeployRequestInfoWithDefaults()
	redeployBody.UseCache = &useCache
	redeployBody.SkipBuild = &skipBuild
	_, resp, err := ctx.Client.ServicesApi.ReDeploy(ctx.Context, service).Info(redeployBody).Execute()

	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while redeploying the service `%s`", serviceName),
			err,
			resp,
		)
	}
	log.Infof("Service %s redeployed.", serviceName)
	return nil
}
