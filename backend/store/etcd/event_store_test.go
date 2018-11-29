// +build integration,!race

package etcd

import (
	"context"
	"reflect"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		event := types.FixtureEvent("entity1", "check1")
		ctx := context.WithValue(context.Background(), types.NamespaceKey, event.Entity.Namespace)

		// Set these to nil in order to avoid comparison issues between {} and nil
		event.Check.Labels = nil
		event.Check.Annotations = nil

		// We should receive an empty slice if no results were found
		events, err := store.GetEvents(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, events)
		assert.Equal(t, len(events), 0)

		err = store.UpdateEvent(ctx, event)
		require.NoError(t, err)

		newEv, err := store.GetEventByEntityCheck(ctx, "entity1", "check1")
		require.NoError(t, err)
		if got, want := newEv, event; !reflect.DeepEqual(got, want) {
			t.Errorf("bad event: got %#v, want %#v", got.Check, want.Check)
		}

		events, err = store.GetEvents(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, len(events))
		if got, want := events[0], event; !reflect.DeepEqual(got, want) {
			t.Errorf("bad event: got %v, want %v", got.Check, want.Check)
		}

		// Get all events with wildcards
		ctx = context.WithValue(ctx, types.NamespaceKey, types.NamespaceTypeAll)
		events, err = store.GetEvents(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))

		// Get all events from a missing namespace
		ctx = context.WithValue(ctx, types.NamespaceKey, "acme")
		events, err = store.GetEvents(ctx)
		require.NoError(t, err)
		require.Equal(t, 0, len(events))

		// Set back the context
		ctx = context.WithValue(ctx, types.NamespaceKey, event.Entity.Namespace)

		newEv, err = store.GetEventByEntityCheck(ctx, "", "foo")
		assert.Nil(t, newEv)
		assert.Error(t, err)

		newEv, err = store.GetEventByEntityCheck(ctx, "foo", "")
		assert.Nil(t, newEv)
		assert.Error(t, err)

		newEv, err = store.GetEventByEntityCheck(ctx, "foo", "foo")
		assert.Nil(t, newEv)
		assert.Nil(t, err)

		events, err = store.GetEventsByEntity(ctx, "entity1")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))
		if got, want := events[0], event; !reflect.DeepEqual(got, want) {
			t.Errorf("bad event: got %v, want %v", got, want)
		}

		assert.NoError(t, store.DeleteEventByEntityCheck(ctx, "entity1", "check1"))
		newEv, err = store.GetEventByEntityCheck(ctx, "entity1", "check1")
		assert.Nil(t, newEv)
		assert.NoError(t, err)

		assert.Error(t, store.DeleteEventByEntityCheck(ctx, "", ""))
		assert.Error(t, store.DeleteEventByEntityCheck(ctx, "", "foo"))
		assert.Error(t, store.DeleteEventByEntityCheck(ctx, "foo", ""))

		// Updating an event in a nonexistent org and env should not work
		event.Entity.Namespace = "missing"
		err = store.UpdateEvent(ctx, event)
		assert.Error(t, err)
	})
}
