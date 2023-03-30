package postgres

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityConfigPoller(t *testing.T) {
	withPostgres(t, func(ctx context.Context, pool *pgxpool.Pool, dsn string) {
		estore := NewEntityConfigStore(pool)
		nsstore := NewNamespaceStore(pool)
		if err := nsstore.CreateIfNotExists(ctx, corev3.FixtureNamespace("default")); err != nil {
			t.Fatal(err)
		}
		if iState, err := estore.List(ctx, "", &store.SelectionPredicate{}); err != nil {
			t.Fatal(err)
		} else if len(iState) > 0 {
			t.Fatalf("unexpected non-empty entities: %v", iState)
		}

		watcherUnderTest := NewWatcher(estore, time.Millisecond*100, time.Millisecond*1500)
		watchCtx, watchCancel := context.WithCancel(ctx)
		req := storev2.ResourceRequest{
			APIVersion: "core/v3",
			Type:       "EntityConfig",
			StoreName:  "entity_configs",
		}
		defer watchCancel()
		watchEvents := watcherUnderTest.Watch(watchCtx, req)

		e1 := corev3.FixtureEntityConfig("test-1")

		// create
		require.NoError(t, estore.CreateIfNotExists(ctx, e1))

		actualEvents := <-watchEvents
		require.Equal(t, 1, len(actualEvents))
		actualEvent := actualEvents[0]
		assert.Equal(t, storev2.WatchCreate, actualEvent.Type)
		value := &corev3.EntityConfig{}
		assert.NoError(t, actualEvent.Value.UnwrapInto(value))
		purgeIndeterminateStoreLabels(value)
		assert.Equal(t, e1, value)

		// update
		e1.Subscriptions = append(e1.Subscriptions, "testingpoller")
		require.NoError(t, estore.UpdateIfExists(ctx, e1))

		actualEvents = <-watchEvents
		require.Equal(t, 1, len(actualEvents))
		actualEvent = actualEvents[0]
		assert.Equal(t, storev2.WatchUpdate, actualEvent.Type)
		assert.NoError(t, actualEvent.Value.UnwrapInto(value))
		purgeIndeterminateStoreLabels(value)
		assert.Equal(t, e1, value)

		// delete
		require.NoError(t, estore.Delete(ctx, "default", "test-1"))

		actualEvents = <-watchEvents
		require.Equal(t, 1, len(actualEvents))
		actualEvent = actualEvents[0]
		assert.Equal(t, storev2.WatchDelete, actualEvent.Type)
		assert.NoError(t, actualEvent.Value.UnwrapInto(value))
		purgeIndeterminateStoreLabels(value)
		assert.Equal(t, e1, value)

		// create multiple
		e2 := corev3.FixtureEntityConfig("test-2")
		require.NoError(t, estore.CreateIfNotExists(ctx, e1))
		require.NoError(t, estore.CreateIfNotExists(ctx, e2))

		twoEvents := make([]storev2.WatchEvent, 0, 2)
		for len(twoEvents) < 2 {
			twoEvents = append(twoEvents, <-watchEvents...)
		}
		assert.Equal(t, 2, len(twoEvents))

		sort.Slice(twoEvents, func(i, j int) bool { return twoEvents[i].Key.Name < twoEvents[j].Key.Name })
		actual1, actual2 := &corev3.EntityConfig{}, &corev3.EntityConfig{}
		twoEvents[0].Value.UnwrapInto(actual1)
		twoEvents[1].Value.UnwrapInto(actual2)

		purgeIndeterminateStoreLabels(actual1)
		purgeIndeterminateStoreLabels(actual2)
		assert.Equal(t, e1, actual1)
		assert.Equal(t, e2, actual2)
	})
}
