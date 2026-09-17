package idmapper

import (
	"context"
	"testing"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPoolMapperResolveIDFullUUIDIsPassedThroughWithoutFetch(t *testing.T) {
	// A nil client proves the UUID path never performs any API call.
	mapper := NewPoolMapper(context.Background(), nil)

	uuid := "123e4567-e89b-42d3-a456-426614174000"
	resolved, err := mapper.ResolveID(uuid)
	require.NoError(t, err)
	assert.Equal(t, uuid, resolved)
}

func TestPoolMapperResolveIDShortIDAndNameUseFetchedMaps(t *testing.T) {
	mapper := NewPoolMapper(context.Background(), nil)
	mapper.fetched = true
	mapper.sidMap.Set("123e4567-e89b-42d3-a456-426614174000", "123e4567")
	mapper.nameMap.Set("123e4567-e89b-42d3-a456-426614174000", "my-pool")

	resolved, err := mapper.ResolveID("123e4567")
	require.NoError(t, err)
	assert.Equal(t, "123e4567-e89b-42d3-a456-426614174000", resolved)

	resolved, err = mapper.ResolveID("my-pool")
	require.NoError(t, err)
	assert.Equal(t, "123e4567-e89b-42d3-a456-426614174000", resolved)
}

func TestPoolMapperResolveIDUnknownValueReturnsCLIError(t *testing.T) {
	mapper := NewPoolMapper(context.Background(), nil)
	mapper.fetched = true

	_, err := mapper.ResolveID("unknown")
	require.Error(t, err)
	var cliErr *errors.CLIError
	require.ErrorAs(t, err, &cliErr)
	assert.Contains(t, cliErr.Error(), "pool")
}
