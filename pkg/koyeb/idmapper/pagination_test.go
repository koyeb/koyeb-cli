package idmapper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchAllPages(t *testing.T) {
	t.Run("terminates when the server ignores the offset", func(t *testing.T) {
		// The smoke repro: the same non-empty page for every offset with a
		// consistent total. Pagination must stop on the reported total, not
		// wait for an empty page that never comes.
		calls := 0
		pools, err := FetchAllPages(func(_, _ int64) ([]string, int64, error) {
			calls++
			return []string{"pool-1"}, 1, nil
		}, 100)
		require.NoError(t, err)
		assert.Equal(t, []string{"pool-1"}, pools)
		assert.Equal(t, 1, calls, "the reported total must end the pagination")
	})

	t.Run("pages until the reported total is covered", func(t *testing.T) {
		calls := 0
		pools, err := FetchAllPages(func(offset, limit int64) ([]string, int64, error) {
			calls++
			switch calls {
			case 1:
				return []string{"p1", "p2"}, 3, nil
			default:
				require.Equal(t, int64(2), offset)
				return []string{"p3"}, 3, nil
			}
		}, 2)
		require.NoError(t, err)
		assert.Equal(t, []string{"p1", "p2", "p3"}, pools)
		assert.Equal(t, 2, calls)
	})

	t.Run("empty page ends the pagination", func(t *testing.T) {
		calls := 0
		pools, err := FetchAllPages(func(_, _ int64) ([]string, int64, error) {
			calls++
			return nil, 5, nil
		}, 100)
		require.NoError(t, err)
		assert.Empty(t, pools)
		assert.Equal(t, 1, calls, "the empty page is the normal exit")
	})

	t.Run("zero total with items still terminates", func(t *testing.T) {
		// A server reporting count=0 while serving items must not trap the
		// CLI either: the offset outgrows any total.
		calls := 0
		pools, err := FetchAllPages(func(_, _ int64) ([]string, int64, error) {
			calls++
			return []string{"p1"}, 0, nil
		}, 100)
		require.NoError(t, err)
		assert.Equal(t, []string{"p1"}, pools)
		assert.Equal(t, 1, calls)
	})

	t.Run("fetch errors propagate", func(t *testing.T) {
		_, err := FetchAllPages(func(_, _ int64) ([]string, int64, error) {
			return nil, 0, fmt.Errorf("boom")
		}, 100)
		require.EqualError(t, err, "boom")
	})
}
