package koyeb

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Create(ctx *CLIContext, cmd *cobra.Command, args []string, createService *koyeb.CreateService) error {
	service, err := h.createService(ctx, cmd, args, createService)
	if err != nil {
		return err
	}
	defer renderServiceState(ctx, cmd, service.GetId())

	if wait, _ := cmd.Flags().GetBool("wait"); wait {
		return waitForServiceDeployment(ctx, cmd, service.GetId())
	}
	return nil
}

// createService resolves the app and creates the service via the API,
// without waiting. Callers own any post-create waiting and rendering.
func (h *ServiceHandler) createService(
	ctx *CLIContext,
	cmd *cobra.Command,
	args []string,
	createService *koyeb.CreateService,
) (*koyeb.Service, error) {
	if err := setProjectHeader(ctx, cmd); err != nil {
		return nil, err
	}
	appID, err := h.parseAppName(cmd, args[0])
	if err != nil {
		return nil, err
	}

	app, err := h.ResolveAppArgs(ctx, appID)
	if err != nil {
		return nil, err
	}

	resApp, resp, err := ctx.Client.AppsApi.GetApp(ctx.Context, app).Execute()
	if err != nil {
		return nil, errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the application `%s`", appID),
			err,
			resp,
		)
	}

	createService.SetAppId(resApp.App.GetId())
	res, resp, err := ctx.Client.ServicesApi.CreateService(ctx.Context).Service(*createService).Execute()
	if err != nil {
		return nil, errors.NewCLIErrorFromAPIError(
			"Error while creating the service",
			err,
			resp,
		)
	}
	log.Infof(
		"Service deployment in progress. To access the build logs, run: `koyeb service logs %s -t build`. For the runtime logs, run `koyeb service logs %s`",
		res.Service.GetId()[:8],
		res.Service.GetId()[:8],
	)
	return res.Service, nil
}

// renderServiceState fetches the service and renders its current state —
// used after create and after an optional wait to show the final status.
func renderServiceState(ctx *CLIContext, cmd *cobra.Command, serviceID string) {
	res, _, err := ctx.Client.ServicesApi.GetService(ctx.Context, serviceID).Execute()
	if err != nil {
		return
	}
	full := GetBoolFlags(cmd, "full")
	getServiceReply := NewGetServiceReply(ctx.Mapper, &koyeb.GetServiceReply{Service: res.Service}, full)
	ctx.Renderer.Render(getServiceReply)
}

// waitForServiceDeployment polls the service until it reaches a steady state
// or --wait-timeout elapses. Unlike the SDK-parity sandbox wait, the
// services wait keeps its historical fail-open classification
// (deploymentWaitDone below).
func waitForServiceDeployment(ctx *CLIContext, cmd *cobra.Command, serviceID string) error {
	waitTimeout, err := waitTimeoutFlag(cmd)
	if err != nil {
		return err
	}
	ctxd, cancel := context.WithTimeout(ctx.Context, waitTimeout)
	defer cancel()

	for range ticker(ctxd, waitPollInterval(cmd)) {
		res, resp, err := ctx.Client.ServicesApi.GetService(ctxd, serviceID).Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				"Error while fetching service",
				err,
				resp,
			)
		}

		if res.Service != nil && res.Service.Status != nil {
			if done, failed := deploymentWaitDone(*res.Service.Status); done {
				if failed {
					return fmt.Errorf("service %s deployment ended in status: %s", serviceID[:8], *res.Service.Status)
				}
				return nil
			}
		}
	}

	log.Infof("Service deployment still in progress, --wait timed out. To access the build logs, run: `koyeb service logs %s -t build`. For the runtime logs, run `koyeb service logs %s`",
		serviceID[:8], serviceID[:8],
	)
	return fmt.Errorf("service deployment still in progress, --wait timed out. To access the build logs, run: `koyeb service logs %s -t build`. For the runtime logs, run `koyeb service logs %s`",
		serviceID[:8], serviceID[:8],
	)
}

// waitTimeoutFlag returns --wait-timeout, rejecting non-positive values
// that would otherwise produce a nonsensical immediate timeout.
func waitTimeoutFlag(cmd *cobra.Command) (time.Duration, error) {
	waitTimeout, _ := cmd.Flags().GetDuration("wait-timeout")
	if waitTimeout <= 0 {
		return 0, &errors.CLIError{
			What:     "Invalid --wait-timeout",
			Why:      "--wait-timeout must be a positive duration",
			Orig:     nil,
			Solution: "Pass a positive duration, e.g. --wait-timeout 5m",
		}
	}
	return waitTimeout, nil
}

// waitPollInterval returns the --wait polling interval: --poll-interval when
// the command registers it (sandbox create), 2s otherwise (services).
// Unusable values fall back to the registered flag default — NaN would
// panic time.NewTicker and Inf would never tick.
func waitPollInterval(cmd *cobra.Command) time.Duration {
	f := cmd.Flags().Lookup("poll-interval")
	if f == nil {
		return 2 * time.Second
	}
	seconds, err := cmd.Flags().GetFloat64("poll-interval")
	if err != nil || seconds <= 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		seconds, _ = strconv.ParseFloat(f.DefValue, 64)
	}
	return time.Duration(seconds * float64(time.Second))
}

// deploymentWaitDone reports whether a --wait polling loop should stop for
// status, and whether that status means the deployment failed. It is the
// historical services classification (fail-open on unknown statuses);
// sandbox and claim waits use the SDKs' fail-closed classifyServiceStatus
// instead — do not consolidate the two.
func deploymentWaitDone(status koyeb.ServiceStatus) (done, failed bool) {
	switch status {
	case koyeb.SERVICESTATUS_DELETED, koyeb.SERVICESTATUS_DEGRADED, koyeb.SERVICESTATUS_UNHEALTHY:
		return true, true
	case koyeb.SERVICESTATUS_STARTING, koyeb.SERVICESTATUS_RESUMING, koyeb.SERVICESTATUS_DELETING, koyeb.SERVICESTATUS_PAUSING:
		return false, false
	default:
		return true, false
	}
}
