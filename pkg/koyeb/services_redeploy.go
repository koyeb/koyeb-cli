package koyeb

import (
	"fmt"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) ReDeploy(ctx *CLIContext, cmd *cobra.Command, args []string) error {

	if err := setProjectHeader(ctx, cmd); err != nil {
		return err
	}
	serviceName, err := h.parseServiceName(cmd, args[0])
	if err != nil {
		return err
	}

	service, err := h.ResolveServiceArgs(ctx, serviceName)
	if err != nil {
		return err
	}

	useCache := GetBoolFlags(cmd, "use-cache")
	skipBuild := GetBoolFlags(cmd, "skip-build")
	wait := GetBoolFlags(cmd, "wait")

	redeployBody := *koyeb.NewRedeployRequestInfoWithDefaults()
	redeployBody.UseCache = &useCache
	redeployBody.SkipBuild = &skipBuild
	res, resp, err := ctx.Client.ServicesApi.ReDeploy(ctx.Context, service).Info(redeployBody).Execute()

	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while redeploying the service `%s`", serviceName),
			err,
			resp,
		)
	}
	log.Infof("Service deployment in progress. To access the build logs, run: `koyeb deployment logs %s -t build`. For the runtime logs, run `koyeb deployment logs %s`",
		res.Deployment.GetId()[:8],
		res.Deployment.GetId()[:8],
	)

	if wait {
		waitTimeout, err := waitTimeoutFlag(cmd)
		if err != nil {
			return err
		}
		if err := waitEngine(ctx.Context, waitTimeout, 2*time.Second,
			deploymentUpdateProbe(ctx, res.Deployment.GetId()),
			func() error { return serviceWaitTimedOut(res.Deployment.GetId(), "deployment logs") },
		); err != nil {
			return err
		}
	}

	log.Infof("Service %s redeployed.", serviceName)
	return nil
}
