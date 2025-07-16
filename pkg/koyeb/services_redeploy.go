package koyeb

import (
	"context"
	"fmt"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) ReDeploy(ctx *CLIContext, cmd *cobra.Command, args []string) error {
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
		ctxd, cancel := context.WithTimeout(ctx.Context, 5*time.Minute)
		defer cancel()

		for range ticker(ctxd, 2*time.Second) {
			res, resp, err := ctx.Client.DeploymentsApi.GetDeployment(ctxd, res.Deployment.GetId()).Execute()
			if err != nil {
				return errors.NewCLIErrorFromAPIError(
					"Error while fetching deployment",
					err,
					resp,
				)
			}

			if res.Deployment != nil && res.Deployment.Status != nil &&
				*res.Deployment.Status != koyeb.DEPLOYMENTSTATUS_ALLOCATING &&
				*res.Deployment.Status != koyeb.DEPLOYMENTSTATUS_PROVISIONING &&
				*res.Deployment.Status != koyeb.DEPLOYMENTSTATUS_PENDING &&
				*res.Deployment.Status != koyeb.DEPLOYMENTSTATUS_STARTING {
				return nil
			}
		}

		log.Infof("Service deployment still in progress, --wait timed out. To access the build logs, run: `koyeb deployment logs %s -t build`. For the runtime logs, run `koyeb deployment logs %s`",
			res.Deployment.GetId()[:8],
			res.Deployment.GetId()[:8],
		)
		return nil
	}

	log.Infof("Service %s redeployed.", serviceName)
	return nil
}
