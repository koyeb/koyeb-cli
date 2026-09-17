package koyeb

import (
	"testing"
	"uuid"

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
