// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventFilterStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		filter := types.FixtureEventFilter("filter1")
		ctx := context.WithValue(context.Background(), types.NamespaceKey, filter.Namespace)

		// We should receive an empty slice if no results were found
		filters, err := store.GetEventFilters(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, filters)

		err = store.UpdateEventFilter(ctx, filter)
		assert.NoError(t, err)

		retrieved, err := store.GetEventFilterByName(ctx, "filter1")
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, filter.Name, retrieved.Name)
		assert.Equal(t, filter.Action, retrieved.Action)
		assert.Equal(t, filter.Expressions, retrieved.Expressions)

		filters, err = store.GetEventFilters(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, filters)
		assert.Equal(t, 1, len(filters))

		// Updating a filter in a nonexistent namespace
		filter.Namespace = "missing"
		err = store.UpdateEventFilter(ctx, filter)
		assert.Error(t, err)
	})
}
