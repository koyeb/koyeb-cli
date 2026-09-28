package koyeb

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWaitEngine(t *testing.T) {
	t.Run("first success wins", func(t *testing.T) {
		calls := 0
		err := waitEngine(context.Background(), time.Second, 5*time.Millisecond,
			func(context.Context) (bool, error) { calls++; return true, nil },
			func() error { return errors.New("must not be reached") })
		require.NoError(t, err)
		assert.Equal(t, 1, calls)
	})

	t.Run("fatal errors surface immediately", func(t *testing.T) {
		boom := errors.New("boom")
		calls := 0
		err := waitEngine(context.Background(), time.Second, 5*time.Millisecond,
			func(context.Context) (bool, error) { calls++; return false, boom },
			func() error { return errors.New("must not be reached") })
		require.ErrorIs(t, err, boom)
		assert.Equal(t, 1, calls)
	})

	t.Run("continue polls until the timeout", func(t *testing.T) {
		calls := 0
		err := waitEngine(context.Background(), 25*time.Millisecond, 5*time.Millisecond,
			func(context.Context) (bool, error) { calls++; return false, nil },
			func() error { return errors.New("timed out") })
		require.EqualError(t, err, "timed out")
		assert.Greater(t, calls, 1)
	})

	t.Run("huge poll intervals never postpone the timeout", func(t *testing.T) {
		start := time.Now()
		err := waitEngine(context.Background(), 20*time.Millisecond, 24*365*time.Hour,
			func(context.Context) (bool, error) { return false, nil },
			func() error { return errors.New("timed out") })
		require.EqualError(t, err, "timed out")
		assert.Less(t, time.Since(start), 5*time.Second)
	})

	t.Run("outer cancellation propagates", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := waitEngine(ctx, time.Second, 5*time.Millisecond,
			func(context.Context) (bool, error) { return false, nil },
			func() error { return errors.New("timed out") })
		require.ErrorIs(t, err, context.Canceled)
	})

	t.Run("probe runs on the deadline context", func(t *testing.T) {
		var probeCtx context.Context
		err := waitEngine(context.Background(), time.Second, 5*time.Millisecond,
			func(c context.Context) (bool, error) { probeCtx = c; return true, nil },
			func() error { return errors.New("timed out") })
		require.NoError(t, err)
		deadline, ok := probeCtx.Deadline()
		require.True(t, ok, "the probe must receive a context bounded by the timeout")
		assert.WithinDuration(t, time.Now().Add(time.Second), deadline, 2*time.Second)
	})
}

func TestDeploymentUpdateWaitDone(t *testing.T) {
	tests := []struct {
		status koyeb.DeploymentStatus
		done   bool
		failed bool
	}{
		{koyeb.DEPLOYMENTSTATUS_HEALTHY, true, false},
		{koyeb.DEPLOYMENTSTATUS_ERROR, true, true},
		{koyeb.DEPLOYMENTSTATUS_DEGRADED, true, true},
		{koyeb.DEPLOYMENTSTATUS_UNHEALTHY, true, true},
		{koyeb.DEPLOYMENTSTATUS_CANCELED, true, true},
		{koyeb.DEPLOYMENTSTATUS_STOPPED, true, true},
		{koyeb.DEPLOYMENTSTATUS_ERRORING, true, true},
		{koyeb.DEPLOYMENTSTATUS_STARTING, false, false},
		{koyeb.DEPLOYMENTSTATUS_PENDING, false, false},
		{koyeb.DEPLOYMENTSTATUS_PROVISIONING, false, false},
		{koyeb.DEPLOYMENTSTATUS_ALLOCATING, false, false},
		// Unknown forward-compat statuses are done-and-healthy (fail-open).
		{koyeb.DeploymentStatus("UNKNOWN_FORWARD_COMPAT"), true, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			done, failed := deploymentUpdateWaitDone(tt.status)
			assert.Equal(t, tt.done, done)
			assert.Equal(t, tt.failed, failed)
		})
	}
}

func TestFailClosedServiceProbe(t *testing.T) {
	t.Run("ready statuses succeed", func(t *testing.T) {
		for _, status := range []koyeb.ServiceStatus{koyeb.SERVICESTATUS_HEALTHY, koyeb.SERVICESTATUS_DEGRADED} {
			probe := failClosedServiceProbe(
				func(context.Context, string) (koyeb.ServiceStatus, error) { return status, nil },
				"svc", func(koyeb.ServiceStatus) error { return errors.New("no") })
			done, err := probe(context.Background())
			require.NoError(t, err)
			assert.True(t, done)
		}
	})

	t.Run("terminal statuses fail via the caller's error", func(t *testing.T) {
		probe := failClosedServiceProbe(
			func(context.Context, string) (koyeb.ServiceStatus, error) { return koyeb.SERVICESTATUS_DELETED, nil },
			"svc", func(koyeb.ServiceStatus) error { return errors.New("terminal") })
		done, err := probe(context.Background())
		require.Error(t, err)
		assert.False(t, done)
	})

	t.Run("transient fetch errors continue", func(t *testing.T) {
		probe := failClosedServiceProbe(
			func(context.Context, string) (koyeb.ServiceStatus, error) { return "", errors.New("blip") },
			"svc", func(koyeb.ServiceStatus) error { return errors.New("no") })
		done, err := probe(context.Background())
		require.NoError(t, err)
		assert.False(t, done)
	})
}
