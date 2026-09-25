package koyeb

import (
	"fmt"

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
// (deploymentWaitDone).
func waitForServiceDeployment(ctx *CLIContext, cmd *cobra.Command, serviceID string) error {
	waitTimeout, err := waitTimeoutFlag(cmd)
	if err != nil {
		return err
	}
	return waitEngine(ctx.Context, waitTimeout, waitPollInterval(cmd),
		serviceDeploymentProbe(ctx, serviceID),
		func() error { return serviceWaitTimedOut(serviceID, "service logs") },
	)
}
