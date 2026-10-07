package koyeb

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"uuid"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	// DefaultClaimWaitTimeout and DefaultClaimPollInterval mirror the
	// Python SDK's wait_claim_ready defaults.
	DefaultClaimWaitTimeout  = 300 * time.Second
	DefaultClaimPollInterval = 2 * time.Second

	// Default claim retry policy, mirroring the Python SDK reference:
	// retry 429/5xx only, at most 3 attempts, linear backoff.
	DefaultClaimAttempts   = 3
	DefaultClaimRetryDelay = 1 * time.Second
)

// claimCaller is the seam the retry loop drives: the API port's Claim in
// production, a fake in tests.
type claimCaller func(ctx context.Context, req koyeb.PoolClaimRequest) (
	*koyeb.PoolClaimReply, *http.Response, error)

// claimRetryable reports whether an API failure with the given HTTP
// status is worth retrying: claims are idempotent per request ID, so
// only transient throttling (429) and server-side failures (5xx)
// qualify. Everything else fails immediately.
func claimRetryable(status int) bool {
	return status == http.StatusTooManyRequests || status >= http.StatusInternalServerError
}

// claimWithRetry runs the claim with the Python reference retry policy:
// HTTP 429/5xx are retried up to maxAttempts attempts with a linear
// retryDelay×attempt backoff. The request — and with it the request ID —
// is reused verbatim across attempts, so a retried claim never consumes
// two pool members.
func claimWithRetry(ctx context.Context, req koyeb.PoolClaimRequest,
	maxAttempts int, retryDelay time.Duration, call claimCaller) (
	*koyeb.PoolClaimReply, *http.Response, error) {
	for attempt := 1; ; attempt++ {
		res, resp, err := call(ctx, req)
		if err == nil {
			return res, resp, nil
		}
		if attempt >= maxAttempts || resp == nil || !claimRetryable(resp.StatusCode) {
			return res, resp, err
		}
		timer := time.NewTimer(retryDelay * time.Duration(attempt))
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, nil, ctx.Err()
		case <-timer.C:
		}
	}
}

// claimRetryPolicy reads and validates the claim retry flags.
func claimRetryPolicy(cmd *cobra.Command) (maxAttempts int, retryDelay time.Duration, err error) {
	maxAttempts, _ = cmd.Flags().GetInt("max-attempts")
	if maxAttempts < 1 {
		return 0, 0, &errors.CLIError{
			What: "Error while claiming an instance from the pool",
			Why:  "the --max-attempts flag must be at least 1",
			Additional: []string{
				"An attempt budget of 1 disables retries; use higher values only for transient 429/5xx throttling.",
			},
			Orig:     nil,
			Solution: "Fix the --max-attempts flag and try again",
		}
	}
	retryDelay, _ = cmd.Flags().GetDuration("retry-delay")
	if retryDelay < 0 {
		return 0, 0, &errors.CLIError{
			What: "Error while claiming an instance from the pool",
			Why:  "the --retry-delay flag cannot be negative",
			Additional: []string{
				"The retry delay uses duration format (e.g. '1s', '500ms'); the wait is --retry-delay × attempt.",
			},
			Orig:     nil,
			Solution: "Fix the --retry-delay flag and try again",
		}
	}
	return maxAttempts, retryDelay, nil
}

// serviceStatusGetter fetches a service status for the shared wait loop.
type serviceStatusGetter func(ctx context.Context, serviceID string) (koyeb.ServiceStatus, error)

// serviceStatusFromClient builds a GetService-backed getter; fetch errors
// are transient input the shared wait loop retries until its timeout.
func serviceStatusFromClient(ctx *CLIContext) serviceStatusGetter {
	return func(c context.Context, id string) (koyeb.ServiceStatus, error) {
		res, _, err := ctx.API.GetService(c, id)
		if err != nil {
			return "", err
		}
		service := res.GetService()
		if !service.HasStatus() {
			return "", fmt.Errorf("service %s has no status", id)
		}
		return service.GetStatus(), nil
	}
}

// serviceStatusClass classifies a service status for claim readiness.
type serviceStatusClass int

const (
	serviceStatusInProgress serviceStatusClass = iota
	serviceStatusReady
	serviceStatusTerminal
)

// classifyServiceStatus fails closed like the SDKs: HEALTHY/DEGRADED are
// ready, STARTING/RESUMING are in progress, everything else is terminal.
// The services wait keeps its fail-open deploymentWaitDone — do not merge.
func classifyServiceStatus(status koyeb.ServiceStatus) serviceStatusClass {
	switch status {
	case koyeb.SERVICESTATUS_HEALTHY, koyeb.SERVICESTATUS_DEGRADED:
		return serviceStatusReady
	case koyeb.SERVICESTATUS_STARTING, koyeb.SERVICESTATUS_RESUMING:
		return serviceStatusInProgress
	default:
		return serviceStatusTerminal
	}
}

// waitClaimReady polls the claimed service until it is ready, mirroring
// the SDKs' wait_claim_ready: transient GetService failures are treated as
// in progress and retried until the timeout, terminal states error out.
func waitClaimReady(ctx context.Context, serviceID string, timeout, pollInterval time.Duration,
	getStatus serviceStatusGetter,
) error {
	return waitEngine(ctx, timeout, pollInterval,
		failClosedServiceProbe(getStatus, serviceID, func(status koyeb.ServiceStatus) error {
			return &errors.CLIError{
				What:     "Claimed service reached a terminal state",
				Why:      fmt.Sprintf("Service '%s' reached terminal state '%s' and will not become ready.", serviceID, status),
				Solution: errors.CLIErrorSolution("Check the service logs with `koyeb service logs " + serviceID + "`"),
			}
		}),
		func() error {
			return &errors.CLIError{
				What:     "Timed out waiting for the claimed service",
				Why:      fmt.Sprintf("service %s did not become ready within %s", serviceID, timeout),
				Solution: errors.CLIErrorSolution("Check the service status with `koyeb service get " + serviceID + "`"),
			}
		},
	)
}

// generateRequestID returns a UUID v7 used when --request-id is not given.
func generateRequestID() string {
	return uuid.NewV7().String()
}

// buildPoolClaimRequest builds the PoolClaimRequest sent to the claim API.
func buildPoolClaimRequest(poolID, requestID string) koyeb.PoolClaimRequest {
	return koyeb.PoolClaimRequest{
		PoolId:    &poolID,
		RequestId: &requestID,
	}
}

func newPoolClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim POOL",
		Short: "Claim an instance from a pool",
		Long: `Claim an instance from a pool.

When --request-id is not provided, a random UUID v7 is generated. Replaying
a claim with the same request ID is idempotent: the server returns the
previously created claim instead of provisioning a new instance. The
request ID is reused across the internal retries.

Retries follow the python SDK reference: only HTTP 429 and 5xx
responses are retried, at most --max-attempts times (default 3), with a
linear --retry-delay × attempt backoff (default 1s). Permanent failures
(family 4xx other than 429) fail immediately.

With --wait, the readiness poll runs until the claimed service is ready,
bounded by --wait-timeout (default 5m) at --poll-interval (default 2s).`,
		Args: cobra.ExactArgs(1),
		Example: `
# Claim an instance from a pool
$> koyeb pool claim my-pool

# Claim with an explicit request ID (idempotent retries)
$> koyeb pool claim my-pool --request-id my-request-id

# Claim with a larger retry budget for heavy throttling
$> koyeb pool claim my-pool --max-attempts 5 --retry-delay 2s
`,
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			maxAttempts, retryDelay, err := claimRetryPolicy(cmd)
			if err != nil {
				return err
			}

			poolID, err := ResolvePoolArgs(ctx, args[0])
			if err != nil {
				return err
			}

			requestID, _ := cmd.Flags().GetString("request-id")
			if requestID == "" {
				requestID = generateRequestID()
			}

			req := buildPoolClaimRequest(poolID, requestID)

			res, resp, err := claimWithRetry(ctx.Context, req, maxAttempts, retryDelay, ctx.API.Claim)
			if err != nil {
				return errors.NewCLIErrorFromAPIError(
					fmt.Sprintf("Error while claiming an instance from the pool `%s`", args[0]),
					err,
					resp,
				)
			}

			return claimWaitFlow(ctx, cmd, res)
		}),
	}
	cmd.Flags().String("request-id", "", "Claim request ID (defaults to a generated UUID v7)")
	cmd.Flags().Int("max-attempts", DefaultClaimAttempts,
		"Max claim attempts on retryable failures, HTTP 429/5xx")
	cmd.Flags().Duration("retry-delay", DefaultClaimRetryDelay,
		"Base delay between claim retries; the wait is --retry-delay × attempt")
	cmd.Flags().Bool("wait", false, "Wait until the claimed service is ready")
	cmd.Flags().Duration("wait-timeout", DefaultClaimWaitTimeout, "Duration the --wait will last until timeout")
	cmd.Flags().Float64("poll-interval", DefaultClaimPollInterval.Seconds(),
		"Seconds between readiness polls when --wait is set")

	return cmd
}

// claimWaitFlow renders the claim and waits for the claimed service when
// --wait is set.
func claimWaitFlow(ctx *CLIContext, cmd *cobra.Command, res *koyeb.PoolClaimReply) error {
	full := GetBoolFlags(cmd, "full")
	claimReply := NewClaimReply(ctx.Mapper, res, full)
	ctx.Renderer.Render(claimReply)

	if !GetBoolFlags(cmd, "wait") {
		return nil
	}

	serviceID := res.GetServiceId()
	if serviceID == "" {
		log.Warnf("Claim reply has no service ID; skipping --wait")
		return nil
	}

	waitTimeout, err := waitTimeoutFlag(cmd)
	if err != nil {
		return err
	}

	if err := waitClaimedService(ctx, serviceID, waitTimeout, waitPollInterval(cmd)); err != nil {
		return err
	}
	log.Infof("Claimed service %s is ready", serviceID)
	return nil
}

// waitClaimedService polls GetService until the claimed service is ready,
// with the --wait-timeout and --poll-interval values from the command.
func waitClaimedService(ctx *CLIContext, serviceID string, timeout, pollInterval time.Duration) error {
	return waitClaimReady(
		ctx.Context, serviceID,
		timeout, pollInterval,
		serviceStatusFromClient(ctx))
}
