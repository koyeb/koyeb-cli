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
	waitTimeout := GetDurationFlags(cmd, "wait-timeout")

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
		ctxd, cancel := context.WithTimeout(ctx.Context, waitTimeout)
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

			if res.Deployment != nil && res.Deployment.Status != nil {
				switch status := *res.Deployment.Status; status {
				case koyeb.DEPLOYMENTSTATUS_ERROR, koyeb.DEPLOYMENTSTATUS_DEGRADED, koyeb.DEPLOYMENTSTATUS_UNHEALTHY, koyeb.DEPLOYMENTSTATUS_CANCELED, koyeb.DEPLOYMENTSTATUS_STOPPED, koyeb.DEPLOYMENTSTATUS_ERRORING:
					return fmt.Errorf("deployment %s update ended in status: %s", res.Deployment.GetId()[:8], status)
				case koyeb.DEPLOYMENTSTATUS_STARTING, koyeb.DEPLOYMENTSTATUS_PENDING, koyeb.DEPLOYMENTSTATUS_PROVISIONING, koyeb.DEPLOYMENTSTATUS_ALLOCATING:
					break
				default:
					return nil
				}
			}
		}

		log.Infof("Service deployment still in progress, --wait timed out. To access the build logs, run: `koyeb deployment logs %s -t build`. For the runtime logs, run `koyeb deployment logs %s`",
			res.Deployment.GetId()[:8],
			res.Deployment.GetId()[:8],
		)
		return fmt.Errorf("service deployment still in progress, --wait timed out. To access the build logs, run: `koyeb deployment logs %s -t build`. For the runtime logs, run `koyeb deployment logs %s`",
			res.Deployment.GetId()[:8],
			res.Deployment.GetId()[:8],
		)
	}

	log.Infof("Service %s redeployed.", serviceName)
	return nil
}
