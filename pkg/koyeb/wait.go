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

// waitProbe runs one poll: done reports whether the wait is satisfied; a
// non-nil error aborts the wait (fatal fetch errors and classified failures
// alike). Transient fetch errors return (false, nil) to retry.
type waitProbe func(ctx context.Context) (done bool, err error)

// waitEngine polls probe until it is satisfied, fails terminally, the timeout
// elapses, or the parent context is cancelled. It never sleeps past the
// deadline, so a huge poll interval cannot postpone the timeout.
func waitEngine(ctx context.Context, timeout, pollInterval time.Duration,
	probe waitProbe,
	timeoutErr func() error,
) error {
	ctxd, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	deadline := time.Now().Add(timeout)

	for {
		done, err := probe(ctxd)
		if err != nil {
			return err
		}
		if done {
			return nil
		}

		if !time.Now().Before(deadline) {
			return timeoutErr()
		}

		sleepFor := pollInterval
		if remaining := time.Until(deadline); remaining < sleepFor {
			sleepFor = remaining
		}
		select {
		case <-time.After(sleepFor):
		case <-ctxd.Done():
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return timeoutErr()
		}
	}
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

// serviceWaitTimedOut logs and returns the shared --wait timeout error.
func serviceWaitTimedOut(id, logCmd string) error {
	// The log line is capitalized; the error keeps its historical lowercase.
	log.Infof(
		"Service deployment still in progress, --wait timed out. "+
			"To access the build logs, run: `koyeb %s %s -t build`. "+
			"For the runtime logs, run `koyeb %s %s`",
		logCmd, id[:8], logCmd, id[:8],
	)
	return fmt.Errorf(
		"service deployment still in progress, --wait timed out. "+
			"To access the build logs, run: `koyeb %s %s -t build`. "+
			"For the runtime logs, run `koyeb %s %s`",
		logCmd, id[:8], logCmd, id[:8],
	)
}

// deploymentWaitDone reports the historical service-status classification for
// --wait loops (fail-open on unknown statuses); the SDKs' fail-closed
// classifyServiceStatus drives the sandbox and claim waits instead — do not
// consolidate the two.
func deploymentWaitDone(status koyeb.ServiceStatus) (done, failed bool) {
	switch status {
	case koyeb.SERVICESTATUS_DELETED, koyeb.SERVICESTATUS_DEGRADED, koyeb.SERVICESTATUS_UNHEALTHY:
		return true, true
	case koyeb.SERVICESTATUS_STARTING, koyeb.SERVICESTATUS_RESUMING,
		koyeb.SERVICESTATUS_DELETING, koyeb.SERVICESTATUS_PAUSING:
		return false, false
	default:
		return true, false
	}
}

// deploymentUpdateWaitDone classifies a deployment status during an update
// wait: error states fail, pre-ready states are in progress, everything else
// counts as success (fail-open, like the service wait).
func deploymentUpdateWaitDone(status koyeb.DeploymentStatus) (done, failed bool) {
	switch status {
	case koyeb.DEPLOYMENTSTATUS_ERROR, koyeb.DEPLOYMENTSTATUS_DEGRADED, koyeb.DEPLOYMENTSTATUS_UNHEALTHY,
		koyeb.DEPLOYMENTSTATUS_CANCELED, koyeb.DEPLOYMENTSTATUS_STOPPED, koyeb.DEPLOYMENTSTATUS_ERRORING:
		return true, true
	case koyeb.DEPLOYMENTSTATUS_STARTING, koyeb.DEPLOYMENTSTATUS_PENDING, koyeb.DEPLOYMENTSTATUS_PROVISIONING,
		koyeb.DEPLOYMENTSTATUS_ALLOCATING:
		return false, false
	default:
		return true, false
	}
}

// serviceDeploymentProbe polls a service's status with the historical
// fail-open classification; fetch errors are fatal.
func serviceDeploymentProbe(ctx *CLIContext, serviceID string) waitProbe {
	return func(c context.Context) (bool, error) {
		res, resp, err := ctx.API.GetService(c, serviceID)
		if err != nil {
			return false, errors.NewCLIErrorFromAPIError(
				"Error while fetching service",
				err,
				resp,
			)
		}
		if res.Service == nil || res.Service.Status == nil {
			return false, nil
		}
		done, failed := deploymentWaitDone(*res.Service.Status)
		if failed {
			return false, fmt.Errorf("service %s deployment ended in status: %s", serviceID[:8], *res.Service.Status)
		}
		return done, nil
	}
}

// deploymentUpdateProbe polls a deployment with the update classification;
// fetch errors are fatal.
func deploymentUpdateProbe(ctx *CLIContext, deploymentID string) waitProbe {
	return func(c context.Context) (bool, error) {
		res, resp, err := ctx.API.GetDeployment(c, deploymentID)
		if err != nil {
			return false, errors.NewCLIErrorFromAPIError(
				"Error while fetching deployment",
				err,
				resp,
			)
		}
		if res.Deployment == nil || res.Deployment.Status == nil {
			return false, nil
		}
		done, failed := deploymentUpdateWaitDone(*res.Deployment.Status)
		if failed {
			return false, fmt.Errorf(
				"deployment %s update ended in status: %s",
				res.Deployment.GetId()[:8], *res.Deployment.Status)
		}
		return done, nil
	}
}

// failClosedServiceProbe builds a probe with the SDKs' fail-closed
// classification; transient fetch errors retry until the timeout.
func failClosedServiceProbe(
	getStatus serviceStatusGetter,
	serviceID string,
	terminalErr func(koyeb.ServiceStatus) error,
) waitProbe {
	return func(c context.Context) (bool, error) {
		status, err := getStatus(c, serviceID)
		if err != nil {
			return false, nil
		}
		switch classifyServiceStatus(status) {
		case serviceStatusReady:
			return true, nil
		case serviceStatusTerminal:
			return false, terminalErr(status)
		}
		return false, nil
	}
}
