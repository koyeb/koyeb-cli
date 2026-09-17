package koyeb

import (
	"context"
	"testing"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func poolClaimFixture() koyeb.PoolClaim {
	id := "123e4567-e89b-42d3-a456-426614174000"
	poolID := "223e4567-e89b-42d3-a456-426614174000"
	serviceID := "323e4567-e89b-42d3-a456-426614174000"
	requestID := "req-123"
	status := koyeb.POOLCLAIMSTATUS_FULFILLED
	createdAt := time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)
	fulfilledAt := time.Date(2024, 5, 1, 12, 1, 0, 0, time.UTC)

	return koyeb.PoolClaim{
		Id:          &id,
		PoolId:      &poolID,
		ServiceId:   &serviceID,
		RequestId:   &requestID,
		Status:      &status,
		CreatedAt:   &createdAt,
		FulfilledAt: &fulfilledAt,
	}
}

func TestClaimReplyHeadersAndFields(t *testing.T) {
	claimID := "123e4567-e89b-42d3-a456-426614174000"
	serviceID := "323e4567-e89b-42d3-a456-426614174000"
	prewarmed := true
	value := &koyeb.PoolClaimReply{
		ClaimId:   &claimID,
		ServiceId: &serviceID,
		Prewarmed: &prewarmed,
	}

	reply := NewClaimReply(idmapper.NewMapper(context.Background(), nil), value, false)

	assert.Equal(t, "Claim", reply.Title())
	assert.Equal(t, []string{"claim_id", "service_id", "prewarmed"}, reply.Headers())

	fields := reply.Fields()
	require.Len(t, fields, 1)
	assert.Equal(t, claimID, fields[0]["claim_id"])
	assert.Equal(t, "323e4567", fields[0]["service_id"])
	assert.Equal(t, "true", fields[0]["prewarmed"])
}

func TestClaimReplyFieldsPrewarmedFalse(t *testing.T) {
	value := koyeb.NewPoolClaimReply()

	reply := NewClaimReply(idmapper.NewMapper(context.Background(), nil), value, false)

	fields := reply.Fields()
	require.Len(t, fields, 1)
	assert.Equal(t, "", fields[0]["claim_id"])
	assert.Equal(t, "", fields[0]["service_id"])
	assert.Equal(t, "false", fields[0]["prewarmed"])
}

func TestClaimReplyMarshalBinary(t *testing.T) {
	value := koyeb.NewPoolClaimReply()
	reply := NewClaimReply(idmapper.NewMapper(context.Background(), nil), value, false)

	data, err := reply.MarshalBinary()
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(data))
}

func TestListPoolClaimsReplyHeadersAndFields(t *testing.T) {
	claim := poolClaimFixture()
	value := &koyeb.ListPoolClaimReply{Claims: []koyeb.PoolClaim{claim}}

	reply := NewListPoolClaimsReply(idmapper.NewMapper(context.Background(), nil), value, false)

	assert.Equal(t, "Claims", reply.Title())
	assert.Equal(t, []string{"id", "request_id", "status", "service_id", "created_at"}, reply.Headers())

	fields := reply.Fields()
	require.Len(t, fields, 1)
	assert.Equal(t, "123e4567-e89b-42d3-a456-426614174000", fields[0]["id"])
	assert.Equal(t, claim.GetRequestId(), fields[0]["request_id"])
	assert.Equal(t, "FULFILLED", fields[0]["status"])
	assert.Equal(t, "323e4567", fields[0]["service_id"])
	assert.NotEmpty(t, fields[0]["created_at"])
}

func TestListPoolClaimsReplyEmptyAndNilSafe(t *testing.T) {
	value := &koyeb.ListPoolClaimReply{}

	reply := NewListPoolClaimsReply(idmapper.NewMapper(context.Background(), nil), value, false)

	assert.Empty(t, reply.Fields())
}

func TestGetPoolClaimReplyHeadersAndFields(t *testing.T) {
	claim := poolClaimFixture()
	value := &koyeb.GetPoolClaimReply{Claim: &claim}

	reply := NewGetPoolClaimReply(idmapper.NewMapper(context.Background(), nil), value, false)

	assert.Equal(t, "Claim", reply.Title())
	assert.Equal(t,
		[]string{"id", "pool_id", "request_id", "status", "service_id", "created_at", "fulfilled_at"},
		reply.Headers())

	fields := reply.Fields()
	require.Len(t, fields, 1)
	assert.Equal(t, "123e4567-e89b-42d3-a456-426614174000", fields[0]["id"])
	assert.Equal(t, "223e4567", fields[0]["pool_id"])
	assert.Equal(t, claim.GetRequestId(), fields[0]["request_id"])
	assert.Equal(t, "FULFILLED", fields[0]["status"])
	assert.Equal(t, "323e4567", fields[0]["service_id"])
	assert.NotEmpty(t, fields[0]["created_at"])
	assert.NotEmpty(t, fields[0]["fulfilled_at"])
}

func TestGetPoolClaimReplyNilSafe(t *testing.T) {
	value := &koyeb.GetPoolClaimReply{Claim: nil}

	reply := NewGetPoolClaimReply(idmapper.NewMapper(context.Background(), nil), value, false)

	fields := reply.Fields()
	require.Len(t, fields, 1)
	assert.Equal(t, "", fields[0]["id"])
	assert.Equal(t, "01 Jan 01 00:00 UTC", fields[0]["created_at"])
	assert.Equal(t, "01 Jan 01 00:00 UTC", fields[0]["fulfilled_at"])
}

func TestGetPoolClaimReplyMarshalBinary(t *testing.T) {
	claim := poolClaimFixture()
	value := &koyeb.GetPoolClaimReply{Claim: &claim}
	reply := NewGetPoolClaimReply(idmapper.NewMapper(context.Background(), nil), value, false)

	data, err := reply.MarshalBinary()
	require.NoError(t, err)
	assert.Contains(t, string(data), `"id":"123e4567-e89b-42d3-a456-426614174000"`)
}
