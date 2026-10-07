package koyeb

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
	"uuid"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRequestID(t *testing.T) {
	seen := make(map[string]bool, 100)

	for i := 0; i < 100; i++ {
		id := generateRequestID()

		assert.False(t, seen[id], "request IDs must be unique")
		seen[id] = true

		_, err := uuid.Parse(id)
		require.NoError(t, err, "generated request ID must be a valid UUID")
	}
}

func TestBuildPoolClaimRequest(t *testing.T) {
	tests := []struct {
		name      string
		poolID    string
		requestID string
	}{
		{"both fields set", "123e4567-e89b-42d3-a456-426614174000", "req-123"},
		{"generated request ID", "123e4567-e89b-42d3-a456-426614174000", generateRequestID()},
		{"empty request ID", "123e4567-e89b-42d3-a456-426614174000", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := buildPoolClaimRequest(tt.poolID, tt.requestID)

			assert.Equal(t, tt.poolID, req.GetPoolId())
			assert.Equal(t, tt.requestID, req.GetRequestId())
		})
	}
}

func TestPoolClaimCmdRegistersWaitFlag(t *testing.T) {
	cmd := newPoolClaimCmd()

	waitFlag := cmd.Flags().Lookup("wait")
	require.NotNil(t, waitFlag, "pool claim must register --wait")
	assert.Equal(t, "false", waitFlag.DefValue, "--wait is opt-in")

	attemptsFlag := cmd.Flags().Lookup("max-attempts")
	require.NotNil(t, attemptsFlag, "pool claim must register --max-attempts")
	assert.Equal(t, "3", attemptsFlag.DefValue, "--max-attempts defaults to 3")

	retryDelayFlag := cmd.Flags().Lookup("retry-delay")
	require.NotNil(t, retryDelayFlag, "pool claim must register --retry-delay")
	assert.Equal(t, "1s", retryDelayFlag.DefValue, "--retry-delay defaults to 1s")

	waitTimeout, err := waitTimeoutFlag(cmd)
	require.NoError(t, err, "pool claim must register --wait-timeout")
	assert.Equal(t, DefaultClaimWaitTimeout, waitTimeout, "--wait-timeout defaults to 5m")

	assert.Equal(t, DefaultClaimPollInterval, waitPollInterval(cmd),
		"--poll-interval defaults to 2s")
	require.NoError(t, cmd.Flags().Set("poll-interval", "0.05"))
	assert.Equal(t, 50*time.Millisecond, waitPollInterval(cmd),
		"--poll-interval is honored when set")
}

func TestClaimRetryPolicy(t *testing.T) {
	t.Run("defaults match the python reference", func(t *testing.T) {
		maxAttempts, retryDelay, err := claimRetryPolicy(newPoolClaimCmd())
		require.NoError(t, err)
		assert.Equal(t, DefaultClaimAttempts, maxAttempts)
		assert.Equal(t, DefaultClaimRetryDelay, retryDelay)
	})

	t.Run("tuning is accepted", func(t *testing.T) {
		cmd := newPoolClaimCmd()
		require.NoError(t, cmd.Flags().Set("max-attempts", "5"))
		require.NoError(t, cmd.Flags().Set("retry-delay", "250ms"))

		maxAttempts, retryDelay, err := claimRetryPolicy(cmd)
		require.NoError(t, err)
		assert.Equal(t, 5, maxAttempts)
		assert.Equal(t, 250*time.Millisecond, retryDelay)
	})

	t.Run("zero attempts are rejected", func(t *testing.T) {
		cmd := newPoolClaimCmd()
		require.NoError(t, cmd.Flags().Set("max-attempts", "0"))

		_, _, err := claimRetryPolicy(cmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--max-attempts flag must be at least 1")
	})

	t.Run("negative delays are rejected", func(t *testing.T) {
		cmd := newPoolClaimCmd()
		require.NoError(t, cmd.Flags().Set("retry-delay", "-1s"))

		_, _, err := claimRetryPolicy(cmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--retry-delay flag cannot be negative")
	})
}

func TestClaimRetryable(t *testing.T) {
	for _, status := range []int{429, 500, 502, 503} {
		assert.True(t, claimRetryable(status), "HTTP %d must be retryable", status)
	}
	for _, status := range []int{200, 400, 401, 404, 409, 422} {
		assert.False(t, claimRetryable(status), "HTTP %d must not be retryable", status)
	}
}

type claimCallResult struct {
	status int
	reply  *koyeb.PoolClaimReply
	err    error
}

// scriptedClaimCaller returns a caller serving the scripted results in
// order (the last one repeats) and records the request of every attempt.
func scriptedClaimCaller(results []claimCallResult, requests *[]koyeb.PoolClaimRequest) claimCaller {
	return func(_ context.Context, req koyeb.PoolClaimRequest) (
		*koyeb.PoolClaimReply, *http.Response, error) {
		*requests = append(*requests, req)
		i := len(*requests) - 1
		if i >= len(results) {
			i = len(results) - 1
		}
		result := results[i]
		if result.err != nil {
			if result.status == 0 {
				return nil, nil, result.err
			}
			return nil, &http.Response{StatusCode: result.status}, result.err
		}
		return result.reply, nil, nil
	}
}

func TestClaimWithRetry(t *testing.T) {
	claimID := "claim-1"
	serviceID := "323e4567-e89b-42d3-a456-426614174000"
	success := &koyeb.PoolClaimReply{ClaimId: &claimID, ServiceId: &serviceID}

	t.Run("succeeds on the first attempt", func(t *testing.T) {
		requests := []koyeb.PoolClaimRequest{}
		res, _, err := claimWithRetry(context.Background(), buildPoolClaimRequest("pool-1", "req-1"),
			3, time.Millisecond, scriptedClaimCaller(
				[]claimCallResult{{reply: success}}, &requests))
		require.NoError(t, err)
		assert.Equal(t, success, res)
		require.Len(t, requests, 1)
	})

	t.Run("retries 429s with linear backoff and preserves the request ID", func(t *testing.T) {
		requests := []koyeb.PoolClaimRequest{}
		start := time.Now()
		res, _, err := claimWithRetry(context.Background(), buildPoolClaimRequest("pool-1", "req-1"),
			3, 5*time.Millisecond, scriptedClaimCaller([]claimCallResult{
				{status: 429, err: fmt.Errorf("too many requests")},
				{status: 429, err: fmt.Errorf("too many requests")},
				{reply: success},
			}, &requests))
		require.NoError(t, err)
		assert.Equal(t, success, res)

		// Three attempts, all carrying the identical request ID.
		require.Len(t, requests, 3)
		for _, req := range requests {
			assert.Equal(t, "req-1", req.GetRequestId())
			assert.Equal(t, "pool-1", req.GetPoolId())
		}
		// Linear backoff: 5ms × 1 + 5ms × 2.
		assert.GreaterOrEqual(t, time.Since(start), 15*time.Millisecond)
	})

	t.Run("retries 5xx", func(t *testing.T) {
		requests := []koyeb.PoolClaimRequest{}
		res, _, err := claimWithRetry(context.Background(), buildPoolClaimRequest("pool-1", "req-1"),
			3, time.Millisecond, scriptedClaimCaller([]claimCallResult{
				{status: 503, err: fmt.Errorf("unavailable")},
				{reply: success},
			}, &requests))
		require.NoError(t, err)
		assert.Equal(t, success, res)
		require.Len(t, requests, 2)
	})

	t.Run("permanent failures fail immediately", func(t *testing.T) {
		requests := []koyeb.PoolClaimRequest{}
		_, resp, err := claimWithRetry(context.Background(), buildPoolClaimRequest("pool-1", "req-1"),
			3, time.Millisecond, scriptedClaimCaller([]claimCallResult{
				{status: 404, err: fmt.Errorf("pool not found")},
			}, &requests))
		require.Error(t, err)
		assert.Equal(t, 404, resp.StatusCode)
		require.Len(t, requests, 1, "404s are never retried")
	})

	t.Run("transport failures fail immediately", func(t *testing.T) {
		requests := []koyeb.PoolClaimRequest{}
		_, resp, err := claimWithRetry(context.Background(), buildPoolClaimRequest("pool-1", "req-1"),
			3, time.Millisecond, scriptedClaimCaller([]claimCallResult{
				{status: 0, err: fmt.Errorf("connection refused")},
			}, &requests))
		require.Error(t, err)
		assert.Nil(t, resp, "a transport failure carries no response")
		require.Len(t, requests, 1)
	})

	t.Run("gives up after max attempts", func(t *testing.T) {
		requests := []koyeb.PoolClaimRequest{}
		_, resp, err := claimWithRetry(context.Background(), buildPoolClaimRequest("pool-1", "req-1"),
			2, time.Millisecond, scriptedClaimCaller([]claimCallResult{
				{status: 429, err: fmt.Errorf("too many requests")},
			}, &requests))
		require.Error(t, err)
		assert.Equal(t, 429, resp.StatusCode)
		require.Len(t, requests, 2)
	})

	t.Run("a single attempt disables retries", func(t *testing.T) {
		requests := []koyeb.PoolClaimRequest{}
		_, _, err := claimWithRetry(context.Background(), buildPoolClaimRequest("pool-1", "req-1"),
			1, time.Millisecond, scriptedClaimCaller([]claimCallResult{
				{status: 429, err: fmt.Errorf("too many requests")},
			}, &requests))
		require.Error(t, err)
		require.Len(t, requests, 1)
	})

	t.Run("context cancellation stops the retries", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		requests := []koyeb.PoolClaimRequest{}

		_, _, err := claimWithRetry(ctx, buildPoolClaimRequest("pool-1", "req-1"),
			3, time.Hour, scriptedClaimCaller([]claimCallResult{
				{status: 429, err: fmt.Errorf("too many requests")},
			}, &requests))
		require.ErrorIs(t, err, context.Canceled)
	})
}

func TestClassifyServiceStatus(t *testing.T) {
	tests := []struct {
		status koyeb.ServiceStatus
		want   serviceStatusClass
	}{
		{koyeb.SERVICESTATUS_HEALTHY, serviceStatusReady},
		{koyeb.SERVICESTATUS_DEGRADED, serviceStatusReady},
		{koyeb.SERVICESTATUS_STARTING, serviceStatusInProgress},
		{koyeb.SERVICESTATUS_RESUMING, serviceStatusInProgress},
		{koyeb.SERVICESTATUS_DELETED, serviceStatusTerminal},
		{koyeb.SERVICESTATUS_UNHEALTHY, serviceStatusTerminal},
		{koyeb.SERVICESTATUS_PAUSING, serviceStatusTerminal},
		{koyeb.SERVICESTATUS_PAUSED, serviceStatusTerminal},
		{koyeb.SERVICESTATUS_DELETING, serviceStatusTerminal},
		{koyeb.ServiceStatus("UNKNOWN_FORWARD_COMPAT"), serviceStatusTerminal},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.want, classifyServiceStatus(tt.status))
		})
	}
}

func statusFunc(statuses []koyeb.ServiceStatus, calls *int) func(context.Context, string) (
	koyeb.ServiceStatus, error) {
	return func(context.Context, string) (koyeb.ServiceStatus, error) {
		i := *calls
		(*calls)++
		if i >= len(statuses) {
			return statuses[len(statuses)-1], nil
		}
		return statuses[i], nil
	}
}

func TestWaitClaimReady(t *testing.T) {
	serviceID := "323e4567-e89b-42d3-a456-426614174000"

	t.Run("ready immediately", func(t *testing.T) {
		calls := 0
		err := waitClaimReady(context.Background(), serviceID, time.Second, 5*time.Millisecond,
			statusFunc([]koyeb.ServiceStatus{koyeb.SERVICESTATUS_HEALTHY}, &calls))
		require.NoError(t, err)
		assert.Equal(t, 1, calls)
	})

	t.Run("in progress then ready", func(t *testing.T) {
		calls := 0
		err := waitClaimReady(context.Background(), serviceID, time.Second, 5*time.Millisecond,
			statusFunc([]koyeb.ServiceStatus{
				koyeb.SERVICESTATUS_STARTING,
				koyeb.SERVICESTATUS_RESUMING,
				koyeb.SERVICESTATUS_HEALTHY,
			}, &calls))
		require.NoError(t, err)
		assert.Equal(t, 3, calls)
	})

	t.Run("transient get failures are retried", func(t *testing.T) {
		failures := 0
		getStatus := func(context.Context, string) (koyeb.ServiceStatus, error) {
			failures++
			if failures <= 2 {
				return "", fmt.Errorf("transient blip")
			}
			return koyeb.SERVICESTATUS_HEALTHY, nil
		}

		err := waitClaimReady(context.Background(), serviceID, time.Second, 5*time.Millisecond, getStatus)
		require.NoError(t, err)
		assert.Equal(t, 3, failures)
	})

	t.Run("terminal state errors immediately", func(t *testing.T) {
		calls := 0
		err := waitClaimReady(context.Background(), serviceID, time.Second, 5*time.Millisecond,
			statusFunc([]koyeb.ServiceStatus{koyeb.SERVICESTATUS_STARTING, koyeb.SERVICESTATUS_DELETED}, &calls))
		require.Error(t, err)
		assert.Contains(t, err.Error(),
			fmt.Sprintf("Service '%s' reached terminal state 'DELETED' and will not become ready.", serviceID))
		assert.Equal(t, 2, calls)
	})

	t.Run("timeout errors", func(t *testing.T) {
		calls := 0
		err := waitClaimReady(context.Background(), serviceID, 20*time.Millisecond, 5*time.Millisecond,
			statusFunc([]koyeb.ServiceStatus{koyeb.SERVICESTATUS_STARTING}, &calls))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "did not become ready")
	})

	t.Run("transient errors are retried until the timeout", func(t *testing.T) {
		calls := 0
		getStatus := func(context.Context, string) (koyeb.ServiceStatus, error) {
			calls++
			return "", fmt.Errorf("still provisioning")
		}

		err := waitClaimReady(context.Background(), serviceID, 20*time.Millisecond, 5*time.Millisecond, getStatus)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "did not become ready")
		assert.Greater(t, calls, 1, "getter must be retried, not failed fast")
	})

	t.Run("context cancellation stops the wait", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := waitClaimReady(ctx, serviceID, time.Second, 5*time.Millisecond,
			func(context.Context, string) (koyeb.ServiceStatus, error) {
				return koyeb.SERVICESTATUS_STARTING, nil
			})
		require.Error(t, err)
	})
}

func TestWaitClaimReadyHugePollIntervalStillTimesOut(t *testing.T) {
	// A saturated poll interval must never sleep past the deadline.
	calls := 0
	start := time.Now()
	err := waitClaimReady(context.Background(), "323e4567-e89b-42d3-a456-426614174000",
		30*time.Millisecond, 24*365*time.Hour,
		func(context.Context, string) (koyeb.ServiceStatus, error) {
			calls++
			return koyeb.SERVICESTATUS_STARTING, nil
		})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "did not become ready")
	assert.Less(t, time.Since(start), 5*time.Second, "the wait must respect the timeout, not the poll interval")
	assert.GreaterOrEqual(t, calls, 2, "the deadline check must still be reached after a poll")
}

func TestClaimWaitFlow(t *testing.T) {
	serviceID := "323e4567-e89b-42d3-a456-426614174000"

	newReply := func(serviceID string) *koyeb.PoolClaimReply {
		return &koyeb.PoolClaimReply{ClaimId: &serviceID, ServiceId: &serviceID}
	}

	t.Run("without --wait the service is not fetched", func(t *testing.T) {
		fake := &fakeAPI{}
		cmd := newPoolClaimCmd()

		require.NoError(t, claimWaitFlow(sandboxTestContext(fake), cmd, newReply(serviceID)))
		assert.Empty(t, fake.servicesFetched)
	})

	t.Run("with --wait the claimed service is awaited", func(t *testing.T) {
		status := koyeb.SERVICESTATUS_HEALTHY
		fake := &fakeAPI{service: &koyeb.Service{Id: &serviceID, Status: &status}}
		cmd := newPoolClaimCmd()
		require.NoError(t, cmd.Flags().Set("wait", "true"))

		require.NoError(t, claimWaitFlow(sandboxTestContext(fake), cmd, newReply(serviceID)))
		assert.Equal(t, []string{serviceID}, fake.servicesFetched)
	})

	t.Run("terminal wait states propagate", func(t *testing.T) {
		status := koyeb.SERVICESTATUS_DELETED
		fake := &fakeAPI{service: &koyeb.Service{Id: &serviceID, Status: &status}}
		cmd := newPoolClaimCmd()
		require.NoError(t, cmd.Flags().Set("wait", "true"))

		err := claimWaitFlow(sandboxTestContext(fake), cmd, newReply(serviceID))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reached terminal state")
	})

	t.Run("empty service ID skips the wait", func(t *testing.T) {
		fake := &fakeAPI{}
		cmd := newPoolClaimCmd()
		require.NoError(t, cmd.Flags().Set("wait", "true"))

		require.NoError(t, claimWaitFlow(sandboxTestContext(fake), cmd, koyeb.NewPoolClaimReply()))
		assert.Empty(t, fake.servicesFetched, "a claim reply without service ID must skip --wait")
	})

	t.Run("--wait-timeout bounds the readiness wait", func(t *testing.T) {
		status := koyeb.SERVICESTATUS_STARTING
		serviceID := "323e4567-e89b-42d3-a456-426614174000"
		fake := &fakeAPI{service: &koyeb.Service{Id: &serviceID, Status: &status}}
		cmd := newPoolClaimCmd()
		require.NoError(t, cmd.Flags().Set("wait", "true"))
		require.NoError(t, cmd.Flags().Set("wait-timeout", "40ms"))

		err := claimWaitFlow(sandboxTestContext(fake), cmd, newReply(serviceID))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "did not become ready within 40ms",
			"the flag value must reach the wait engine")
	})

	t.Run("an invalid --wait-timeout is rejected", func(t *testing.T) {
		fake := &fakeAPI{}
		cmd := newPoolClaimCmd()
		require.NoError(t, cmd.Flags().Set("wait", "true"))
		require.NoError(t, cmd.Flags().Set("wait-timeout", "0"))

		err := claimWaitFlow(sandboxTestContext(fake), cmd, newReply("323e4567-e89b-42d3-a456-426614174000"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--wait-timeout must be a positive duration")
	})

	t.Run("--poll-interval drives the poll cadence", func(t *testing.T) {
		status := koyeb.SERVICESTATUS_STARTING
		serviceID := "323e4567-e89b-42d3-a456-426614174000"
		fake := &fakeAPI{service: &koyeb.Service{Id: &serviceID, Status: &status}}
		cmd := newPoolClaimCmd()
		require.NoError(t, cmd.Flags().Set("wait", "true"))
		require.NoError(t, cmd.Flags().Set("wait-timeout", "150ms"))
		require.NoError(t, cmd.Flags().Set("poll-interval", "0.005"))

		start := time.Now()
		err := claimWaitFlow(sandboxTestContext(fake), cmd, newReply(serviceID))
		elapsed := time.Since(start)

		require.Error(t, err)
		assert.Less(t, elapsed, 2*time.Second, "5ms polls must bound the wait, not the 2s default")
		assert.GreaterOrEqual(t, len(fake.servicesFetched), 3,
			"a 5ms cadence over a 150ms budget must poll repeatedly")
	})
}
