// +build integration,!race

package etcd

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
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
		events, _, err := store.GetEvents(ctx, 0, "")
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

		events, _, err = store.GetEvents(ctx, 0, "")
		require.NoError(t, err)
		require.Equal(t, 1, len(events))
		if got, want := events[0], event; !reflect.DeepEqual(got, want) {
			t.Errorf("bad event: got %v, want %v", got.Check, want.Check)
		}

		// Get all events with wildcards
		ctx = context.WithValue(ctx, corev2.NamespaceKey, corev2.NamespaceTypeAll)
		events, _, err = store.GetEvents(ctx, 0, "")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))

		// Get all events from a missing namespace
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "acme")
		events, _, err = store.GetEvents(ctx, 0, "")
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
		event := corev2.FixtureEvent("entity1", "check1")
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)
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

func TestGetEventsPagination(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		// Create a "testing" namespace in the store
		testingNS := corev2.FixtureNamespace("testing")
		store.UpdateNamespace(context.Background(), testingNS)

		// Add 42 objects in the store: 21 in the "default" namespace and 21 in
		// the "testing" namespace
		for i := 1; i <= 21; i++ {
			// We force the entity and check number to be 2 digits
			// "wide" in order to have a "natural" order: 01, 02, ...
			// instead of 1, 11, ...
			entityName := fmt.Sprintf("entity%.2d", i)
			checkName := fmt.Sprintf("check%.2d", i)

			event := corev2.FixtureEvent(entityName, checkName)
			event.Name = fmt.Sprintf("%s/%s", entityName, checkName)

			if err := store.UpdateEvent(context.Background(), event); err != nil {
				t.Fatal(err)
			}

			event.Namespace = "testing"
			event.Entity.Namespace = "testing"

			if err := store.UpdateEvent(context.Background(), event); err != nil {
				t.Fatal(err)
			}
		}

		// Test that we can retrieve all 42 objects in 8 pages of 5 items
		// and a final page of 2 items, in the expected order: 01 through 21 in
		// namespace "default" then 01 through 21 in namespace "testing"
		ctx := context.Background()
		pageSize := 5
		continueToken := ""

		for i := 0; i < 8; i++ {
			events, nextContinueToken, err := store.GetEvents(ctx, int64(pageSize), continueToken)
			if err != nil {
				t.Fatal(err)
			}

			if len(events) != 5 {
				t.Fatalf("Expected page %d to have 5 items, got %d", i, len(events))
			}

			offset := i * 5
			for j, event := range events {
				n := ((offset + j) % 21) + 1
				expected := fmt.Sprintf("entity%.2d/check%.2d", n, n)

				if event.Name != expected {
					t.Fatalf("Expected %s, got %s (%s)", expected, event.Name, event.Namespace)
				}
			}

			continueToken = nextContinueToken
		}

		// Check the last page (2 items)
		events, nextContinueToken, err := store.GetEvents(ctx, int64(pageSize), continueToken)
		if err != nil {
			t.Fatal(err)
		}

		if len(events) != 2 {
			t.Fatalf("Expected a page with 2 items, got %d", len(events))
		}

		if nextContinueToken != "" {
			t.Fatalf("Expected next continue token to be \"\", got %s", nextContinueToken)
		}

		offset := 40
		for j, event := range events {
			n := ((offset + j) % 21) + 1
			expected := fmt.Sprintf("entity%.2d/check%.2d", n, n)

			if event.Name != expected {
				t.Fatalf("Expected %s, got %s", expected, event.Name)
			}
		}

		// Test that we can limit the query to the "default" namespace
		// This is to make sure that the don't "escape" the namespace when there
		// are more entities stored in a namespace after "default".
		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")
		t.Run("through default namespace", func(t *testing.T) {
			testPagination(t, ctx, store, 10, 21)
		})

		// Test that we can limit the query to the "testing" namespace
		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "testing")
		t.Run("through testing namespace", func(t *testing.T) {
			testPagination(t, ctx, store, 10, 21)
		})

		// Test with limit=1
		ctx = context.Background()
		pageSize = 1
		continueToken = ""

		for i := 0; i < 42; i++ {
			events, nextContinueToken, err := store.GetEvents(ctx, int64(pageSize), continueToken)
			if err != nil {
				t.Fatal(err)
			}

			if len(events) != 1 {
				t.Fatalf("Expected page %d to have 1 items, got %d", i, len(events))
			}

			offset := i * 1
			for j, event := range events {
				n := ((offset + j) % 21) + 1
				expected := fmt.Sprintf("entity%.2d/check%.2d", n, n)

				if event.Name != expected {
					t.Fatalf("Expected %s, got %s (%s)", expected, event.Name, event.Namespace)
				}
			}

			continueToken = nextContinueToken
		}

		// TODO: Add test with limit > setSize
	})
}

func testPagination(t *testing.T, ctx context.Context, etcd store.Store, pageSize, setSize int) {
	nFullPages := setSize / pageSize
	nLeftovers := setSize % pageSize

	continueToken := ""
	for i := 0; i < nFullPages; i++ {
		events, nextContinueToken, err := etcd.GetEvents(ctx, int64(pageSize), continueToken)
		if err != nil {
			t.Fatal(err)
		}

		if len(events) != pageSize {
			t.Fatalf("Expected page %d to have %d items but got %d", i, pageSize, len(events))
		}

		offset := i * pageSize
		for j, event := range events {
			n := ((offset + j) % setSize) + 1
			expected := fmt.Sprintf("entity%.2d/check%.2d", n, n)

			if event.Name != expected {
				t.Fatalf("Expected %s, got %s (namespace=%s)", expected, event.Name, event.Namespace)
			}
		}

		continueToken = nextContinueToken
	}

	// Check the last page, supposed to hold nLeftovers items
	events, nextContinueToken, err := etcd.GetEvents(ctx, int64(pageSize), continueToken)
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != nLeftovers {
		t.Fatalf("Expected last page with %d items, got %d", nLeftovers, len(events))
	}

	if nextContinueToken != "" {
		t.Fatalf("Expected next continue token to be \"\", got %s", nextContinueToken)
	}

	offset := pageSize * nFullPages
	for j, event := range events {
		n := ((offset + j) % setSize) + 1
		expected := fmt.Sprintf("entity%.2d/check%.2d", n, n)

		if event.Name != expected {
			t.Fatalf("Expected %s, got %s", expected, event.Name)
		}
	}
}

func TestGetEventsByEntityPagination(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		// Create a "testing" namespace in the store
		testingNS := corev2.FixtureNamespace("testing")
		store.UpdateNamespace(context.Background(), testingNS)

		// Add 42 objects in the store: 21 checks for entity1 in the "default"
		// namespace and 21 checks for "entity1" in the "testing" namespace
		for i := 1; i <= 21; i++ {
			// We force the entity and check number to be 2 digits
			// "wide" in order to have a "natural" order: 01, 02, ...
			// instead of 1, 11, ...
			checkName := fmt.Sprintf("check%.2d", i)

			event := corev2.FixtureEvent("entity1", checkName)
			event.Name = fmt.Sprintf("entity1/%s", checkName)

			if err := store.UpdateEvent(context.Background(), event); err != nil {
				t.Fatal(err)
			}

			event.Namespace = "testing"
			event.Entity.Namespace = "testing"

			if err := store.UpdateEvent(context.Background(), event); err != nil {
				t.Fatal(err)
			}
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, corev2.PageContinueKey, "")
		ctx = context.WithValue(ctx, corev2.PageSizeKey, 10)
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")
		t.Run("entity1 in default namespace", func(t *testing.T) {
			testGetEventsByEntityPagination(t, store, 21, ctx, "entity1")
		})

		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.PageContinueKey, "")
		ctx = context.WithValue(ctx, corev2.PageSizeKey, 10)
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "testing")
		t.Run("entity1 in testing namespace", func(t *testing.T) {
			testGetEventsByEntityPagination(t, store, 21, ctx, "entity1")
		})

		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.PageContinueKey, "")
		ctx = context.WithValue(ctx, corev2.PageSizeKey, 1)
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")
		t.Run("page size equals one", func(t *testing.T) {
			testGetEventsByEntityPagination(t, store, 21, ctx, "entity1")
		})

		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.PageContinueKey, "")
		ctx = context.WithValue(ctx, corev2.PageSizeKey, 1337)
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")
		t.Run("page size bigger than set size", func(t *testing.T) {
			testGetEventsByEntityPagination(t, store, 21, ctx, "entity1")
		})
	})
}

func testGetEventsByEntityPagination(t *testing.T, etcd store.Store, setSize int, ctx context.Context, entityName string) {
	pageSize := store.PageSizeFromContext(ctx)

	nFullPages := setSize / pageSize
	nLeftovers := setSize % pageSize

	for i := 0; i < nFullPages; i++ {
		events, err := etcd.GetEventsByEntity(ctx, entityName)
		if err != nil {
			t.Fatal(err)
		}

		if len(events) != pageSize {
			t.Fatalf("Expected page %d to have %d items but got %d", i, pageSize, len(events))
		}

		offset := i * pageSize
		for j, event := range events {
			n := ((offset + j) % setSize) + 1
			expected := fmt.Sprintf("%s/check%.2d", entityName, n)

			if event.Name != expected {
				t.Fatalf("Expected %s, got %s", expected, event.Name)
			}
		}

		lastItem := events[len(events)-1]
		continueKey := fmt.Sprintf("%s\x00", lastItem.Check.Name)
		ctx = context.WithValue(ctx, corev2.PageContinueKey, continueKey)
	}

	// Check the last page, supposed to hold nLeftovers items
	events, err := etcd.GetEventsByEntity(ctx, entityName)
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != nLeftovers {
		t.Fatalf("Expected last page with %d items, got %d", nLeftovers, len(events))
	}

	offset := pageSize * nFullPages
	for j, event := range events {
		n := ((offset + j) % setSize) + 1
		expected := fmt.Sprintf("%s/check%.2d", entityName, n)

		if event.Name != expected {
			t.Fatalf("Expected %s, got %s", expected, event.Name)
		}
	}
}
