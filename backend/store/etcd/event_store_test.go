// +build integration,!race

package etcd

import (
	"context"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventStorageMaxOutputSize(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		event := corev2.FixtureEvent("entity1", "check1")
		event.Check.Output = "VERY LONG"
		event.Check.MaxOutputSize = 4
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)
		if err := store.UpdateEvent(ctx, event); err != nil {
			t.Fatal(err)
		}
		event, err := store.GetEventByEntityCheck(ctx, "entity1", "check1")
		if err != nil {
			t.Fatal(err)
		}
		if got, want := event.Check.Output, "VERY"; got != want {
			t.Fatalf("bad check output: got %q, want %q", got, want)
		}
	})
}

func TestEventStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		event := corev2.FixtureEvent("entity1", "check1")
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)

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
		ctx = context.WithValue(ctx, corev2.NamespaceKey, corev2.NamespaceTypeAll)
		events, err = store.GetEvents(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))

		// Get all events from a missing namespace
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "acme")
		events, err = store.GetEvents(ctx)
		require.NoError(t, err)
		require.Equal(t, 0, len(events))

		// Set back the context
		ctx = context.WithValue(ctx, corev2.NamespaceKey, event.Entity.Namespace)

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

func TestDoNotStoreMetrics(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		event := corev2.FixtureEvent("entity1", "check1")
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)
		event.Metrics = &corev2.Metrics{
			Handlers: []string{"metrix"},
		}
		if err := store.UpdateEvent(ctx, event); err != nil {
			t.Fatal(err)
		}
		if event, err := store.GetEventByEntityCheck(ctx, event.Entity.Name, event.Check.Name); err != nil {

			t.Fatal(err)
		} else if event.Metrics != nil {
			t.Fatal("expected nil metrics")
		}
	})
}

func TestUpdateEventWithZeroTimestamp_GH2636(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		event := types.FixtureEvent("entity1", "check1")
		ctx := context.WithValue(context.Background(), types.NamespaceKey, event.Entity.Namespace)
		event.Timestamp = 0

		if err := store.UpdateEvent(ctx, event); err != nil {
			t.Fatal(err)
		}

		storedEvent, err := store.GetEventByEntityCheck(ctx, event.Entity.Name, event.Check.Name)
		if err != nil {
			t.Fatal(err)
		}

		if storedEvent.Timestamp == 0 {
			t.Fatal("expected non-zero timestamp")
		}
	})
}
