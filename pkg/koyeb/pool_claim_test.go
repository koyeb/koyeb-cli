package koyeb

import (
	"context"
	"fmt"
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

func statusFunc(statuses []koyeb.ServiceStatus, calls *int) func(context.Context, string) (koyeb.ServiceStatus, error) {
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
			statusFunc([]koyeb.ServiceStatus{koyeb.SERVICESTATUS_STARTING, koyeb.SERVICESTATUS_RESUMING, koyeb.SERVICESTATUS_HEALTHY}, &calls))
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
		assert.Contains(t, err.Error(), fmt.Sprintf("Service '%s' reached terminal state 'DELETED' and will not become ready.", serviceID))
		assert.Equal(t, 2, calls)
	})

	t.Run("timeout errors", func(t *testing.T) {
		calls := 0
		err := waitClaimReady(context.Background(), serviceID, 20*time.Millisecond, 5*time.Millisecond,
			statusFunc([]koyeb.ServiceStatus{koyeb.SERVICESTATUS_STARTING}, &calls))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "did not become ready")
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
