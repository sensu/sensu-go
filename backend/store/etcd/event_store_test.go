// +build integration,!race

package etcd

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestEventStorageMaxOutputSize(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		event := corev2.FixtureEvent("entity1", "check1")
		event.Check.Output = "VERY LONG"
		event.Check.MaxOutputSize = 4
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)
		if _, _, err := store.UpdateEvent(ctx, event); err != nil {
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
	testWithEtcd(t, func(s store.Store) {
		// Create new namespaces
		require.NoError(t, s.CreateNamespace(context.Background(), types.FixtureNamespace("acme")))
		require.NoError(t, s.CreateNamespace(context.Background(), types.FixtureNamespace("acme-devel")))

		event := corev2.FixtureEvent("entity1", "check1")
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)
		pred := &store.SelectionPredicate{}

		// Set these to nil in order to avoid comparison issues between {} and nil
		event.Check.Labels = nil
		event.Check.Annotations = nil

		// We should receive an empty slice if no results were found
		events, err := s.GetEvents(ctx, pred)
		assert.NoError(t, err)
		assert.NotNil(t, events)
		assert.Equal(t, len(events), 0)
		assert.Empty(t, pred.Continue)

		_, _, err = s.UpdateEvent(ctx, event)
		require.NoError(t, err)

		newEv, err := s.GetEventByEntityCheck(ctx, "entity1", "check1")
		require.NoError(t, err)
		if got, want := newEv, event; !reflect.DeepEqual(got, want) {
			t.Errorf("bad event: got %#v, want %#v", got.Check, want.Check)
		}

		events, err = s.GetEvents(ctx, pred)
		require.NoError(t, err)
		require.Equal(t, 1, len(events))
		require.Empty(t, pred.Continue)
		if got, want := events[0], event; !reflect.DeepEqual(got, want) {
			t.Errorf("bad event: got %v, want %v", got.Check, want.Check)
		}

		// Add an event in the acme namespace
		event.Entity.Namespace = "acme"
		ctx = context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)
		_, _, err = s.UpdateEvent(ctx, event)
		require.NoError(t, err)

		// Add an event in the acme-devel namespace
		event.Entity.Namespace = "acme-devel"
		ctx = context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)
		_, _, err = s.UpdateEvent(ctx, event)
		require.NoError(t, err)

		// Get all events with wildcards
		ctx = context.WithValue(ctx, corev2.NamespaceKey, corev2.NamespaceTypeAll)
		events, err = s.GetEvents(ctx, pred)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(events))
		assert.Empty(t, pred.Continue)

		// Get all events in the acme namespace
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "acme")
		events, err = s.GetEvents(ctx, pred)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))
		assert.Empty(t, pred.Continue)

		// Get all events in the acme-devel namespace
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "acme-devel")
		events, err = s.GetEvents(ctx, pred)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))
		assert.Empty(t, pred.Continue)

		// Get all events from a missing namespace
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "missing")
		events, err = s.GetEvents(ctx, pred)
		require.NoError(t, err)
		require.Equal(t, 0, len(events))
		require.Empty(t, pred.Continue)

		// Set back the context
		ctx = context.WithValue(ctx, corev2.NamespaceKey, event.Entity.Namespace)

		newEv, err = s.GetEventByEntityCheck(ctx, "", "foo")
		assert.Nil(t, newEv)
		assert.Error(t, err)

		newEv, err = s.GetEventByEntityCheck(ctx, "foo", "")
		assert.Nil(t, newEv)
		assert.Error(t, err)

		newEv, err = s.GetEventByEntityCheck(ctx, "foo", "foo")
		assert.Nil(t, newEv)
		assert.Nil(t, err)

		events, err = s.GetEventsByEntity(ctx, "entity1", pred)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))
		assert.Empty(t, pred.Continue)
		if got, want := events[0], event; !reflect.DeepEqual(got, want) {
			t.Errorf("bad event: got %v, want %v", got, want)
		}

		assert.NoError(t, s.DeleteEventByEntityCheck(ctx, "entity1", "check1"))
		newEv, err = s.GetEventByEntityCheck(ctx, "entity1", "check1")
		assert.Nil(t, newEv)
		assert.NoError(t, err)

		assert.Error(t, s.DeleteEventByEntityCheck(ctx, "", ""))
		assert.Error(t, s.DeleteEventByEntityCheck(ctx, "", "foo"))
		assert.Error(t, s.DeleteEventByEntityCheck(ctx, "foo", ""))

		// Updating an event in a nonexistent namespace should not work
		event.Entity.Namespace = "missing"
		_, _, err = s.UpdateEvent(ctx, event)
		assert.Error(t, err)
	})
}

func TestEventByEntity(t *testing.T) {
	testWithEtcd(t, func(s store.Store) {
		// Create new namespaces
		require.NoError(t, s.CreateNamespace(context.Background(), types.FixtureNamespace("acme")))

		e1 := corev2.FixtureEvent("entity", "check1")
		e2 := corev2.FixtureEvent("entity1", "check1")
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, e1.Entity.Namespace)
		pred := &store.SelectionPredicate{}

		// Set these to nil in order to avoid comparison issues between {} and nil
		e1.Check.Labels = nil
		e1.Check.Annotations = nil
		e2.Check.Labels = nil
		e2.Check.Annotations = nil

		_, _, err := s.UpdateEvent(ctx, e1)
		require.NoError(t, err)
		_, _, err = s.UpdateEvent(ctx, e2)
		require.NoError(t, err)

		// Listing events for entity should not return the event for entity1
		events, err := s.GetEventsByEntity(ctx, "entity", pred)
		assert.NoError(t, err)
		assert.NotNil(t, events)
		assert.Equal(t, 1, len(events))
		assert.Empty(t, pred.Continue)
		if got, want := events[0], e1; !reflect.DeepEqual(got, want) {
			t.Errorf("bad event: got %#v, want %#v", got.Check, want.Check)
		}

		// Listing events for entity1 should still work even though entity exists
		events, err = s.GetEventsByEntity(ctx, "entity1", pred)
		assert.NoError(t, err)
		assert.NotNil(t, events)
		assert.Equal(t, len(events), 1)
		assert.Empty(t, pred.Continue)
		if got, want := events[0], e2; !reflect.DeepEqual(got, want) {
			t.Errorf("bad event: got %#v, want %#v", got.Check, want.Check)
		}
	})
}

func TestDoNotStoreMetrics(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		event := corev2.FixtureEvent("entity1", "check1")
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)
		event.Metrics = &corev2.Metrics{
			Handlers: []string{"metrix"},
		}
		if _, _, err := store.UpdateEvent(ctx, event); err != nil {
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

		if _, _, err := store.UpdateEvent(ctx, event); err != nil {
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

func TestCheckOccurrences(t *testing.T) {
	OK := uint32(0)
	WARN := uint32(1)
	CRIT := uint32(2)

	// 1. Event with OK Check status - Occurrences: 1, OccurrencesWatermark: 1
	// 2. Event with OK Check status - Occurrences: 2, OccurrencesWatermark: 2
	// 3. Event with WARN Check status - Occurrences: 1, OccurrencesWatermark: 1
	// 4. Event with WARN Check status - Occurrences: 2, OccurrencesWatermark: 2
	// 5. Event with WARN Check status - Occurrences: 3, OccurrencesWatermark: 3
	// 6. Event with CRIT Check status - Occurrences: 1, OccurrencesWatermark: 3
	// 7. Event with CRIT Check status - Occurrences: 2, OccurrencesWatermark: 3
	// 8. Event with CRIT Check status - Occurrences: 3, OccurrencesWatermark: 3
	// 9. Event with CRIT Check status - Occurrences: 4, OccurrencesWatermark: 4
	// 10. Event with OK Check status - Occurrences: 1, OccurrencesWatermark: 4
	// 11. Event with CRIT Check status - Occurrences: 1, OccurrencesWatermark: 1
	testCases := []struct {
		name                         string
		status                       uint32
		expectedOccurrences          int64
		expectedOccurrencesWatermark int64
	}{
		{
			name:                         "OK",
			status:                       OK,
			expectedOccurrences:          1,
			expectedOccurrencesWatermark: 1,
		},
		{
			name:                         "OK -> OK",
			status:                       OK,
			expectedOccurrences:          2,
			expectedOccurrencesWatermark: 2,
		},
		{
			name:                         "OK -> WARN",
			status:                       WARN,
			expectedOccurrences:          1,
			expectedOccurrencesWatermark: 1,
		},
		{
			name:                         "WARN -> WARN",
			status:                       WARN,
			expectedOccurrences:          2,
			expectedOccurrencesWatermark: 2,
		},
		{
			name:                         "WARN -> WARN",
			status:                       WARN,
			expectedOccurrences:          3,
			expectedOccurrencesWatermark: 3,
		},
		{
			name:                         "WARN -> CRIT",
			status:                       CRIT,
			expectedOccurrences:          1,
			expectedOccurrencesWatermark: 3,
		},
		{
			name:                         "CRIT -> CRIT",
			status:                       CRIT,
			expectedOccurrences:          2,
			expectedOccurrencesWatermark: 3,
		},
		{
			name:                         "CRIT -> CRIT",
			status:                       CRIT,
			expectedOccurrences:          3,
			expectedOccurrencesWatermark: 3,
		},
		{
			name:                         "CRIT -> CRIT",
			status:                       CRIT,
			expectedOccurrences:          4,
			expectedOccurrencesWatermark: 4,
		},
		{
			name:                         "CRIT -> OK",
			status:                       OK,
			expectedOccurrences:          1,
			expectedOccurrencesWatermark: 4,
		},
		{
			name:                         "OK -> CRIT",
			status:                       CRIT,
			expectedOccurrences:          1,
			expectedOccurrencesWatermark: 1,
		},
	}

	event := corev2.FixtureEvent("entity1", "check1")
	event.Check.Occurrences = 1
	event.Check.OccurrencesWatermark = 1
	event.Check.History = []corev2.CheckHistory{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event.Check.Status = tc.status
			event.Check.History = append(event.Check.History, corev2.CheckHistory{Status: tc.status})
			updateOccurrences(event.Check)
			assert.Equal(t, tc.expectedOccurrences, event.Check.Occurrences)
			assert.Equal(t, tc.expectedOccurrencesWatermark, event.Check.OccurrencesWatermark)
		})
	}
}

func TestGetEventsPagination(t *testing.T) {
	testWithEtcd(t, func(s store.Store) {
		// Create a "testing" namespace in the store
		testingNS := corev2.FixtureNamespace("testing")
		s.UpdateNamespace(context.Background(), testingNS)

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

			if _, _, err := s.UpdateEvent(context.Background(), event); err != nil {
				t.Fatal(err)
			}

			event.Namespace = "testing"
			event.Entity.Namespace = "testing"

			if _, _, err := s.UpdateEvent(context.Background(), event); err != nil {
				t.Fatal(err)
			}
		}

		// Test that we can retrieve all 42 objects in 8 pages of 5 items
		// and a final page of 2 items, in the expected order: 01 through 21 in
		// namespace "default" then 01 through 21 in namespace "testing"
		ctx := context.Background()
		pred := &store.SelectionPredicate{Limit: 5}

		for i := 0; i < 8; i++ {
			events, err := s.GetEvents(ctx, pred)
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
		}

		// Check the last page (2 items)
		events, err := s.GetEvents(ctx, pred)
		if err != nil {
			t.Fatal(err)
		}

		if len(events) != 2 {
			t.Fatalf("Expected a page with 2 items, got %d", len(events))
		}

		if pred.Continue != "" {
			t.Fatalf("Expected next continue token to be \"\", got %s", pred.Continue)
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
			testPagination(t, ctx, s, 10, 21)
		})

		// Test that we can limit the query to the "testing" namespace
		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "testing")
		t.Run("through testing namespace", func(t *testing.T) {
			testPagination(t, ctx, s, 10, 21)
		})

		// Test with limit=1
		ctx = context.Background()
		pred = &store.SelectionPredicate{Limit: 1}

		for i := 0; i < 42; i++ {
			events, err := s.GetEvents(ctx, pred)
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
		}

		// TODO: Add test with limit > setSize
	})
}

func testPagination(t *testing.T, ctx context.Context, etcd store.Store, pageSize, setSize int) {
	pred := &store.SelectionPredicate{Limit: int64(pageSize)}
	nFullPages := setSize / pageSize
	nLeftovers := setSize % pageSize

	for i := 0; i < nFullPages; i++ {
		events, err := etcd.GetEvents(ctx, pred)
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
	}

	// Check the last page, supposed to hold nLeftovers items
	if nLeftovers > 0 {
		events, err := etcd.GetEvents(ctx, pred)
		if err != nil {
			t.Fatal(err)
		}

		if len(events) != nLeftovers {
			t.Fatalf("Expected last page with %d items, got %d", nLeftovers, len(events))
		}

		if pred.Continue != "" {
			t.Fatalf("Expected next continue token to be \"\", got %s", pred.Continue)
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

			if _, _, err := store.UpdateEvent(context.Background(), event); err != nil {
				t.Fatal(err)
			}

			event.Namespace = "testing"
			event.Entity.Namespace = "testing"

			if _, _, err := store.UpdateEvent(context.Background(), event); err != nil {
				t.Fatal(err)
			}
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")
		t.Run("entity1 in default namespace", func(t *testing.T) {
			testGetEventsByEntityPagination(t, ctx, store, 10, 21, "entity1")
		})

		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "testing")
		t.Run("entity1 in testing namespace", func(t *testing.T) {
			testGetEventsByEntityPagination(t, ctx, store, 10, 21, "entity1")
		})

		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")
		t.Run("page size equals one", func(t *testing.T) {
			testGetEventsByEntityPagination(t, ctx, store, 1, 21, "entity1")
		})

		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")
		t.Run("page size bigger than set size", func(t *testing.T) {
			testGetEventsByEntityPagination(t, ctx, store, 1337, 21, "entity1")
		})
	})
}

func testGetEventsByEntityPagination(t *testing.T, ctx context.Context, etcd store.Store, pageSize, setSize int, entityName string) {
	pred := &store.SelectionPredicate{Limit: int64(pageSize)}
	nFullPages := setSize / pageSize
	nLeftovers := setSize % pageSize

	for i := 0; i < nFullPages; i++ {
		events, err := etcd.GetEventsByEntity(ctx, entityName, pred)
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
	}

	// Check the last page, supposed to hold nLeftovers items
	if nLeftovers > 0 {
		events, err := etcd.GetEventsByEntity(ctx, entityName, pred)
		if err != nil {
			t.Fatal(err)
		}

		if len(events) != nLeftovers {
			t.Fatalf("Expected last page with %d items, got %d", nLeftovers, len(events))
		}

		if pred.Continue != "" {
			t.Fatalf("Expected next continue token to be \"\", got %s", pred.Continue)
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
}

func TestHandleExpireOnResolveEntries(t *testing.T) {
	expireOnResolve := func(s *corev2.Silenced) *corev2.Silenced {
		s.ExpireOnResolve = true
		return s
	}

	resolution := func(e *corev2.Event) *corev2.Event {
		e.Check.History = []corev2.CheckHistory{
			corev2.CheckHistory{Status: 1},
			corev2.CheckHistory{Status: 0},
		}
		e.Check.Status = 0
		return e
	}

	testCases := []struct {
		name                    string
		event                   *corev2.Event
		silencedEntry           *corev2.Silenced
		expectedSilencedEntries []string
	}{
		{
			name:                    "Non-resolution Non-expire-on-resolve Event",
			event:                   corev2.FixtureEvent("entity1", "check1"),
			silencedEntry:           corev2.FixtureSilenced("sub1:check1"),
			expectedSilencedEntries: []string{"sub1:check1"},
		},
		{
			name:                    "Non-Resolution Expire-on-resolve Event",
			event:                   corev2.FixtureEvent("entity1", "check1"),
			silencedEntry:           expireOnResolve(corev2.FixtureSilenced("sub1:check1")),
			expectedSilencedEntries: []string{"sub1:check1"},
		},
		{
			name:                    "Resolution Non-expire-on-resolve Event",
			event:                   resolution(corev2.FixtureEvent("entity1", "check1")),
			silencedEntry:           corev2.FixtureSilenced("sub1:check1"),
			expectedSilencedEntries: []string{"sub1:check1"},
		},
		{
			name:                    "Resolution Expire-on-resolve Event",
			event:                   resolution(corev2.FixtureEvent("entity1", "check1")),
			silencedEntry:           expireOnResolve(corev2.FixtureSilenced("sub1:check1")),
			expectedSilencedEntries: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), corev2.NamespaceKey, "default")

			mockStore := &mockstore.MockStore{}

			mockStore.On(
				"GetSilencedEntriesByName",
				mock.Anything,
				mock.Anything,
			).Return([]*corev2.Silenced{tc.silencedEntry}, nil)

			mockStore.On(
				"DeleteSilencedEntryByName",
				mock.Anything,
				mock.Anything,
			).Return(nil)

			tc.event.Check.Silenced = []string{tc.silencedEntry.Name}

			err := handleExpireOnResolveEntries(ctx, tc.event, mockStore)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedSilencedEntries, tc.event.Check.Silenced)
		})
	}
}

func TestEventStoreHistory(t *testing.T) {
	testWithEtcd(t, func(s store.Store) {
		ctx := store.NamespaceContext(context.Background(), "default")
		event := corev2.FixtureEvent("foo", "bar")
		want := []corev2.CheckHistory{}
		for i := 0; i < 30; i++ {
			event.Check.Executed = int64(i)
			historyItem := corev2.CheckHistory{
				Executed: int64(i),
			}
			want = append(want, historyItem)
			_, _, err := s.UpdateEvent(ctx, event)
			if err != nil {
				t.Fatal(err)
			}
		}
		want = want[len(want)-21:]
		event, err := s.GetEventByEntityCheck(ctx, "foo", "bar")
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 21; i++ {
		}
		if got := event.Check.History; !reflect.DeepEqual(got, want) {
			t.Fatalf("bad event history: got %v, want %v", got, want)
		}
	})
}
