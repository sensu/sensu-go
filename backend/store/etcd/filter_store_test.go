//go:build integration && !race
// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventFilterStorage(t *testing.T) {
	testWithEtcd(t, func(s store.Store) {
		filter := corev2.FixtureEventFilter("filter1")
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, filter.Namespace)

		// We should receive an empty slice if no results were found
		pred := &store.SelectionPredicate{}
		filters, err := s.GetEventFilters(ctx, pred)
		assert.NoError(t, err)
		assert.NotNil(t, filters)
		assert.Empty(t, pred.Continue)

		err = s.UpdateEventFilter(ctx, filter)
		assert.NoError(t, err)

		retrieved, err := s.GetEventFilterByName(ctx, "filter1")
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, filter.Name, retrieved.Name)
		assert.Equal(t, filter.Action, retrieved.Action)
		assert.Equal(t, filter.Expressions, retrieved.Expressions)

		filters, err = s.GetEventFilters(ctx, pred)
		require.NoError(t, err)
		require.NotEmpty(t, filters)
		assert.Equal(t, 1, len(filters))
		assert.Empty(t, pred.Continue)

		// Updating a filter in a nonexistent namespace
		filter.Namespace = "missing"
		err = s.UpdateEventFilter(ctx, filter)
		assert.Error(t, err)
	})
}
