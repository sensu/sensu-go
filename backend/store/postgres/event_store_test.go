package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/selector"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventStorageMaxOutputSize(t *testing.T) {
	testWithPostgresStore(t, func(store store.Store) {
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
	testWithPostgresStore(t, func(s store.Store) {
		event := corev2.FixtureEvent("entity1", "check1")
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)
		pred := &store.SelectionPredicate{}

		// Set these to nil in order to avoid comparison issues between {} and nil
		event.Check.Labels = nil
		event.Check.Annotations = nil

		// Reset this, as the history source of truth is different for postgres
		event.Check.History = []corev2.CheckHistory{
			{
				Status:   event.Check.Status,
				Executed: event.Check.Executed,
			},
		}

		// We should receive an empty slice if no results were found
		events, err := s.GetEvents(ctx, pred)
		assert.NoError(t, err)
		assert.NotNil(t, events)
		assert.Equal(t, len(events), 0)
		assert.Empty(t, pred.Continue)

		_, _, err = s.UpdateEvent(ctx, event)
		require.NoError(t, err)

		// Set state to passing, as we expect the store to handle this for us
		event.Check.State = corev2.EventPassingState

		newEv, err := s.GetEventByEntityCheck(ctx, "entity1", "check1")
		require.NoError(t, err)
		if got, want := newEv.Check, event.Check; !reflect.DeepEqual(got, want) {
			t.Errorf("bad event: got %#v, want %#v", got, want)
		}

		if got, want := newEv.Check.State, corev2.EventPassingState; got != want {
			t.Errorf("bad Check.State: got %q, want %q", got, want)
		}

		events, err = s.GetEvents(ctx, pred)
		require.NoError(t, err)
		require.Equal(t, 1, len(events))
		require.Empty(t, pred.Continue)
		if got, want := events[0].Check, event.Check; !reflect.DeepEqual(got, want) {
			t.Errorf("bad event: got %v, want %v", got, want)
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
		if got, want := events[0].Check, event.Check; !reflect.DeepEqual(got, want) {
			t.Errorf("bad event: got %v, want %v", got, want)
		}

		assert.NoError(t, s.DeleteEventByEntityCheck(ctx, "entity1", "check1"))
		newEv, err = s.GetEventByEntityCheck(ctx, "entity1", "check1")
		assert.NoError(t, err)
		assert.Nil(t, newEv)

		assert.Error(t, s.DeleteEventByEntityCheck(ctx, "", ""))
		assert.Error(t, s.DeleteEventByEntityCheck(ctx, "", "foo"))
		assert.Error(t, s.DeleteEventByEntityCheck(ctx, "foo", ""))

		// Updating an event in a nonexistent namespace should not work
		// TODO(echlebek): reconcile this behaviour with the etcd store
		// event.Entity.Namespace = "missing"
		// err = s.UpdateEvent(ctx, event)
		// assert.Error(t, err)
	})
}

func TestDoNotStoreMetrics(t *testing.T) {
	testWithPostgresStore(t, func(store store.Store) {
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
	testWithPostgresStore(t, func(store store.Store) {
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

func TestGetEventsOrdering(t *testing.T) {
	testWithPostgresStore(t, func(s store.Store) {
		ctx := store.NamespaceContext(context.Background(), "default")

		for i := 0; i < 5; i++ {
			event := corev2.FixtureEvent(fmt.Sprintf("entity%d", i), "check")
			event.Check.LastOK = int64(i)
			event.Check.Status = uint32(i)
			_, _, err := s.UpdateEvent(ctx, event)
			if err != nil {
				t.Fatal(err)
			}
		}

		sp := store.SelectionPredicate{}

		// Test get events ordered by entity ascending
		sp.Ordering = corev2.EventSortEntity
		sp.Descending = false
		events, err := s.GetEvents(ctx, &sp)
		if err != nil {
			t.Fatal(err)
		}

		for i, event := range events {
			if event.Entity.Name != fmt.Sprintf("entity%d", i) {
				t.Error("unexpected ordering by entity ascending")
			}
		}

		// Test get events ordered by entity descending
		sp.Ordering = corev2.EventSortEntity
		sp.Descending = true
		events, err = s.GetEvents(ctx, &sp)
		if err != nil {
			t.Fatal(err)
		}

		for i, event := range events {
			if event.Entity.Name != fmt.Sprintf("entity%d", len(events)-1-i) {
				t.Error("unexpected ordering by entity descending")
			}
		}

		// Test get events ordered by last_ok ascending
		sp.Ordering = corev2.EventSortLastOk
		sp.Descending = false
		events, err = s.GetEvents(ctx, &sp)
		if err != nil {
			t.Fatal(err)
		}

		for i := 1; i < len(events); i++ {
			if events[i-1].Check.LastOK > events[i].Check.LastOK {
				t.Errorf("unexpected result when ordering by last_ok ascending: event %d should be before event %d", i, i-1)
			}
		}

		// Test get events ordered by last_ok descending
		sp.Ordering = corev2.EventSortLastOk
		sp.Descending = true
		events, err = s.GetEvents(ctx, &sp)
		if err != nil {
			t.Fatal(err)
		}

		for i := 1; i < len(events); i++ {
			if events[i-1].Check.LastOK < events[i].Check.LastOK {
				t.Errorf("unexpected result when ordering by last_ok descending: event %d should be before event %d", i, i-1)
			}
		}

		// Test get events ordered by severity ascending
		sp.Ordering = corev2.EventSortSeverity
		sp.Descending = false
		events, err = s.GetEvents(ctx, &sp)
		if err != nil {
			t.Fatal(err)
		}

		deriveSeverity := func(e *corev2.Event) int {
			if e.HasCheck() {
				switch e.Check.Status {
				case 0:
					return 3
				case 1:
					return 1
				case 2:
					return 0
				default:
					return 2
				}
			}
			return 4
		}

		for i := 1; i < len(events); i++ {
			if deriveSeverity(events[i-1]) > deriveSeverity(events[i]) {
				t.Errorf("unexpected result when ordering by severity ascending: event %d should be before event %d", i, i-1)
			}
		}

		// Test get events ordered by severiy descending
		sp.Ordering = corev2.EventSortSeverity
		sp.Descending = true
		events, err = s.GetEvents(ctx, &sp)
		if err != nil {
			t.Fatal(err)
		}

		for i := 1; i < len(events); i++ {
			if deriveSeverity(events[i-1]) < deriveSeverity(events[i]) {
				t.Errorf("unexpected result when ordering by severity descending: event %d should be before event %d", i, i-1)
			}
		}

		// Test get events ordered by timestamp ascending
		sp.Ordering = corev2.EventSortTimestamp
		sp.Descending = false
		events, err = s.GetEvents(ctx, &sp)
		if err != nil {
			t.Fatal(err)
		}

		for i := 1; i < len(events); i++ {
			if events[i-1].Timestamp > events[i].Timestamp {
				t.Errorf("unexpected result when ordering by timestamp ascending: event %d should be before event %d", i, i-1)
			}
		}

		// Test get events ordered by timestamp descending
		sp.Ordering = corev2.EventSortTimestamp
		sp.Descending = true
		events, err = s.GetEvents(ctx, &sp)
		if err != nil {
			t.Fatal(err)
		}

		for i := 1; i < len(events); i++ {
			if events[i-1].Timestamp < events[i].Timestamp {
				t.Errorf("unexpected result when ordering by timestamp descending: event %d should be before event %d", i, i-1)
			}
		}

		// Test that unknown orderings raise an error
		sp.Ordering = "garbage"
		want := "resource is invalid: unknown ordering requested"
		_, err = s.GetEvents(ctx, &sp)
		if err.Error() != want {
			t.Errorf("unexpected error: got '%s', want '%s'", err.Error(), want)
		}
	})
}

func TestGetEventsPagination(t *testing.T) {
	testWithPostgresStore(t, func(s store.Store) {
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
	testWithPostgresStore(t, func(store store.Store) {
		// Create a "testing" namespace in the store
		// testingNS := corev2.FixtureNamespace("testing")
		// store.UpdateNamespace(context.Background(), testingNS)

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

func TestUpdateEventsWithMetrics(t *testing.T) {
	testWithPostgresStore(t, func(store store.Store) {
		event := corev2.FixtureEvent("foo", "bar")
		event.Metrics = &corev2.Metrics{
			Handlers: []string{
				"foo", "bar",
			},
			Points: []*corev2.MetricPoint{
				{
					Name:      "foo",
					Value:     0.5,
					Timestamp: 12345,
				},
			},
		}
		ctx := context.Background()
		updatedEvent, previousEvent, err := store.UpdateEvent(ctx, event)
		if err != nil {
			t.Fatal(err)
		}
		if previousEvent != nil {
			t.Errorf("previous event is not nil")
		}
		if got, want := updatedEvent.Metrics, event.Metrics; !reflect.DeepEqual(got, want) {
			t.Errorf("bad updated metrics: got %#v, want %#v", got, want)
		}
	})
}

func TestUpdateEventHasCheckState(t *testing.T) {
	testWithPostgresStore(t, func(store store.Store) {
		event := corev2.FixtureEvent("foo", "bar")
		ctx := context.Background()
		updatedEvent, previousEvent, err := store.UpdateEvent(ctx, event)
		if err != nil {
			t.Fatal(err)
		}
		if previousEvent != nil {
			t.Errorf("previous event is not nil")
		}
		if got, want := updatedEvent.Check.State, corev2.EventPassingState; got != want {
			t.Fatalf("bad check state: got %q, want %q", got, want)
		}
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

func TestEventStoreHistory(t *testing.T) {
	testWithPostgresStore(t, func(s store.Store) {
		event := corev2.FixtureEvent("foo", "bar")
		ctx := store.NamespaceContext(context.Background(), "default")
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
		want = want[9:]
		if len(want) != 21 {
			t.Fatal("bad want")
		}
		event, err := s.GetEventByEntityCheck(ctx, "foo", "bar")
		if err != nil {
			t.Fatal(err)
		}
		got := event.Check.History
		if got, want := len(got), 21; got != want {
			t.Errorf("bad history length: got %d, want %d", got, want)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("bad event history: got %v, want %v", got, want)
		}
	})
}

func TestEventStoreSelectors(t *testing.T) {
	pgURL := os.Getenv("PG_URL")
	if pgURL == "" {
		pgURL = "host=/run/postgresql sslmode=disable"
	}
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		st, err := NewEventStore(db, nil, Config{
			DSN: pgURL,
		}, 1)
		if err != nil {
			t.Fatal(err)
		}
		event := corev2.FixtureEvent("foo", "bar")
		event.Entity.Labels = map[string]string{"foo": "bar"}
		event.Check.Labels = map[string]string{"foo": "baz"}
		ctx = store.NamespaceContext(ctx, "default")
		if _, _, err := st.UpdateEvent(ctx, event); err != nil {
			t.Fatal(err)
		}
		{
			rows, err := db.Query(ctx, `SELECT selectors->>'event.check.name' FROM events WHERE selectors->'event.check.name' IS NOT NULL;`)
			if err != nil {
				t.Fatal(err)
			}
			for rows.Next() {
				var selector string
				if err := rows.Scan(&selector); err != nil {
					t.Fatal(err)
				}
			}
			rows.Close()
		}
		row := db.QueryRow(ctx, `SELECT count(*) FROM events WHERE selectors @> '{"event.check.status": "0"}'`)
		var count int
		if err := row.Scan(&count); err != nil {
			t.Fatal(err)
		}
		if got, want := count, 1; got != want {
			t.Errorf("bad count: got %d, want %d", got, want)
		}
		row = db.QueryRow(ctx, `SELECT selectors->'event.check.status', selectors->'event.entity.labels.foo' FROM events`)
		var statusB, fooB []byte
		if err := row.Scan(&statusB, &fooB); err != nil {
			t.Fatal(err)
		}
		var status, foo string
		if err := json.Unmarshal(statusB, &status); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(fooB, &foo); err != nil {
			t.Fatal(err)
		}
		if got, want := status, fmt.Sprintf("%d", event.Check.Status); got != want {
			t.Errorf("bad entity class: got %q, want %q", got, want)
		}
		if got, want := foo, "bar"; got != want {
			t.Errorf("bad label value: got %q, want %q", got, want)
		}
	})
}

func TestEventStoreSignedInteger_GH4000(t *testing.T) {
	testWithPostgresStore(t, func(s store.Store) {
		event := corev2.FixtureEvent("foo", "bar")
		status := int32(-1)
		event.Check.Status = uint32(status)
		ctx := store.NamespaceContext(context.Background(), "default")
		_, _, err := s.UpdateEvent(ctx, event)
		if err != nil {
			t.Fatal(err)
		}
		event, err = s.GetEventByEntityCheck(ctx, "foo", "bar")
		if err != nil {
			t.Fatal(err)
		}
		if got, want := int32(event.Check.Status), status; got != want {
			t.Errorf("bad status: got %d, want %d", got, want)
		}
	})
}

func TestEventStatusTransition(t *testing.T) {
	testWithPostgresStore(t, func(s store.Store) {
		ctx := store.NamespaceContext(context.Background(), "default")
		event := corev2.FixtureEvent("foo", "bar")
		if _, _, err := s.UpdateEvent(ctx, event); err != nil {
			t.Fatal(err)
		}
		event.Check.Status = 1
		if _, _, err := s.UpdateEvent(ctx, event); err != nil {
			t.Fatal(err)
		}
	})
}

func TestEventCheckStateSelector(t *testing.T) {
	testWithPostgresStore(t, func(s store.Store) {
		ctx := store.NamespaceContext(context.Background(), "default")
		event := corev2.FixtureEvent("foo", "bar")
		event.Check.LowFlapThreshold = 1
		event.Check.HighFlapThreshold = 10
		event.Check.History = nil
		for i := 0; i < 30; i++ {
			event.Check.Status = uint32(i % 2)
			event.Check.Executed = time.Now().Unix() + int64(i)
			if _, _, err := s.UpdateEvent(ctx, event); err != nil {
				t.Fatal(err)
			}
		}
		selektor := &selector.Selector{
			Operations: []selector.Operation{
				selector.Operation{
					LValue:        "event.check.state",
					Operator:      selector.DoubleEqualSignOperator,
					RValues:       []string{"flapping"},
					OperationType: selector.OperationTypeFieldSelector,
				},
			},
		}
		ctx = selector.ContextWithSelector(context.Background(), selektor)
		events, err := s.GetEvents(ctx, &store.SelectionPredicate{})
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != 1 {
			t.Fatal("no events found")
		}
	})
}

func TestEventOccurrencesRegression_GH1469(t *testing.T) {
	testWithPostgresStore(t, func(s store.Store) {
		ctx := store.NamespaceContext(context.Background(), "default")
		event := corev2.FixtureEvent("foo", "bar")
		event.Check.Status = 1
		for i := 0; i < 10; i++ {
			if _, _, err := s.UpdateEvent(ctx, event); err != nil {
				t.Fatal(err)
			}
		}
		event, err := s.GetEventByEntityCheck(ctx, "foo", "bar")
		if err != nil {
			t.Fatal(err)
		}
		if got, want := event.Check.OccurrencesWatermark, int64(10); got != want {
			t.Errorf("bad occurrences watermark: got %d, want %d", got, want)
		}
	})
}

func TestCountEvents(t *testing.T) {
	testWithPostgresStore(t, func(s store.Store) {
		ctx := store.NamespaceContext(context.Background(), "default")
		event := corev2.FixtureEvent("foo", "bar")
		for i := 0; i < 30; i++ {
			event.Check.Name = fmt.Sprintf("%d", i)
			_, _, err := s.UpdateEvent(ctx, event)
			if err != nil {
				t.Fatal(err)
			}
		}
		count, err := s.CountEvents(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := count, int64(30); got != want {
			t.Errorf("bad event count: got %d, want %d", got, want)
		}
		selektor := &selector.Selector{
			Operations: []selector.Operation{
				selector.Operation{
					LValue:        "event.check.name",
					Operator:      selector.DoubleEqualSignOperator,
					RValues:       []string{"1"},
					OperationType: selector.OperationTypeFieldSelector,
				},
			},
		}
		selCtx := selector.ContextWithSelector(context.Background(), selektor)
		count, err = s.CountEvents(selCtx, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := count, int64(1); got != want {
			t.Errorf("bad count: got %d, want %d", got, want)
		}
		ctx = store.NamespaceContext(context.Background(), "foobar")
		event.Entity.Namespace = "foobar"
		event.Check.Namespace = "foobar"
		for i := 0; i < 15; i++ {
			event.Check.Name = fmt.Sprintf("%d", i)
			_, _, err := s.UpdateEvent(ctx, event)
			if err != nil {
				t.Fatal(err)
			}
		}
		count, err = s.CountEvents(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := count, int64(15); got != want {
			t.Errorf("bad event count: got %d, want %d", got, want)
		}
		ctx = store.NamespaceContext(context.Background(), "notexists")
		count, err = s.CountEvents(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := count, int64(0); got != want {
			t.Errorf("bad event count: got %d, want %d", got, want)
		}
	})
}
