package memory

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

// it's like corev2.FixtureEvent but the entity class is agent
func fixtureEvent(entity, check string) *corev2.Event {
	event := corev2.FixtureEvent(entity, check)
	event.Entity.EntityClass = corev2.EntityAgentClass
	return event
}

// mock.On always appends, so if we want to replace a method, have to do something else.
// Do not use this in a concurrent setting.
func On(mock *mock.Mock, method string, args ...interface{}) *mock.Call {
	for _, call := range mock.ExpectedCalls {
		if call.Method == method {
			return call
		}
	}
	// fall through is regular On()
	return mock.On(method, args...)
}

func TestMemoryWriteTo(t *testing.T) {
	store := new(mockstore.MockStore)
	On(&store.Mock, "UpdateEvent", mock.Anything, mock.Anything).Return((*corev2.Event)(nil), (*corev2.Event)(nil), nil)
	db := newMemoryDB(rate.NewLimiter(1000, 1000))
	event := fixtureEvent("entity", "check")
	eventBytes, err := proto.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	compressedEventBytes := snappy.Encode(nil, eventBytes)
	if err != nil {
		t.Fatal(err)
	}
	entry := db.ReadEntry("default", "entity", "check")
	entry.Dirty = true
	entry.EventBytes = compressedEventBytes
	err = db.WriteTo(context.Background(), store)
	if err != nil {
		t.Fatal(err)
	}
	store.AssertCalled(t, "UpdateEvent", mock.Anything, mock.Anything)
}

func readEvent(db *memorydb, namespace, entity, check string) *corev2.Event {
	entry := db.ReadEntry(namespace, entity, check)
	if entry.EventBytes == nil {
		return nil
	}
	decompressed, err := snappy.Decode(nil, entry.EventBytes)
	if err != nil {
		panic(err)
	}
	var event corev2.Event
	if err := proto.Unmarshal(decompressed, &event); err != nil {
		panic(err)
	}
	return &event
}

func TestEventStorageMaxOutputSize(t *testing.T) {
	ms := new(mockstore.MockStore)
	config := EventStoreConfig{
		BackingStore:    ms,
		FlushInterval:   10 * time.Second,
		SilenceStore:    new(mockstore.MockStore),
		EventWriteLimit: 1000,
	}
	s := NewEventStore(config)
	event := fixtureEvent("entity1", "check1")
	event.Check.Output = "VERY LONG"
	event.Check.MaxOutputSize = 4
	event.Entity.EntityClass = corev2.EntityAgentClass
	ctx := context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)
	On(&ms.Mock, "GetEventByEntityCheck", mock.Anything, "entity1", "check1").Return((*corev2.Event)(nil), &store.ErrNotFound{})
	On(&ms.Mock, "UpdateEvent", mock.Anything, mock.Anything).Return((*corev2.Event)(nil), (*corev2.Event)(nil), nil)
	if _, _, err := s.UpdateEvent(ctx, event); err != nil {
		t.Fatal(err)
	}
	On(&ms.Mock, "GetEventByEntityCheck", mock.Anything, "entity1", "check1").Return(readEvent(s.db, "default", "entity1", "check1"), nil)
	event, err := s.GetEventByEntityCheck(ctx, "entity1", "check1")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := event.Check.Output, "VERY"; got != want {
		t.Fatalf("bad check output: got %q, want %q", got, want)
	}
	ms.AssertCalled(t, "GetEventByEntityCheck", mock.Anything, "entity1", "check1")
}

func TestEventStorage(t *testing.T) {
	ms := new(mockstore.MockStore)
	config := EventStoreConfig{
		BackingStore:    ms,
		FlushInterval:   10 * time.Second,
		SilenceStore:    new(mockstore.MockStore),
		EventWriteLimit: 1000,
	}
	s := NewEventStore(config)
	event := fixtureEvent("entity1", "check1")
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
	On(&ms.Mock, "GetEvents", mock.Anything, mock.Anything).Return(([]*corev2.Event)(nil), nil)
	events, err := s.GetEvents(ctx, pred)
	assert.NoError(t, err)
	assert.Empty(t, events)
	assert.Empty(t, pred.Continue)

	On(&ms.Mock, "GetEventByEntityCheck", mock.Anything, mock.Anything, mock.Anything).Return((*corev2.Event)(nil), &store.ErrNotFound{})
	On(&ms.Mock, "UpdateEvent", mock.Anything, mock.Anything).Return((*corev2.Event)(nil), (*corev2.Event)(nil), nil)
	_, _, err = s.UpdateEvent(ctx, event)
	require.NoError(t, err)

	ms.AssertCalled(t, "GetEventByEntityCheck", mock.Anything, mock.Anything, mock.Anything)

	// Set state to passing, as we expect the store to handle this for us
	event.Check.State = corev2.EventPassingState

	On(&ms.Mock, "GetEventByEntityCheck", mock.Anything, mock.Anything, mock.Anything).Return(readEvent(s.db, "default", "entity1", "check1"), nil)
	newEv, err := s.GetEventByEntityCheck(ctx, "entity1", "check1")
	require.NoError(t, err)
	if got, want := newEv.Check, event.Check; !reflect.DeepEqual(got, want) {
		t.Errorf("bad event: got %#v, want %#v", got, want)
	}

	if got, want := newEv.Check.State, corev2.EventPassingState; got != want {
		t.Errorf("bad Check.State: got %q, want %q", got, want)
	}

	ms.AssertCalled(t, "GetEventByEntityCheck", mock.Anything, mock.Anything, mock.Anything)

	On(&ms.Mock, "GetEvents", mock.Anything, mock.Anything).Return([]*corev2.Event{newEv}, nil)
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
}

func TestDoNotStoreMetrics(t *testing.T) {
	ms := new(mockstore.MockStore)
	config := EventStoreConfig{
		BackingStore:    ms,
		FlushInterval:   10 * time.Second,
		SilenceStore:    new(mockstore.MockStore),
		EventWriteLimit: 1000,
	}
	s := NewEventStore(config)
	event := fixtureEvent("entity1", "check1")
	ctx := context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)
	event.Metrics = &corev2.Metrics{
		Handlers: []string{"metrix"},
	}
	ms.On("GetEventByEntityCheck", mock.Anything, mock.Anything, mock.Anything).Return((*corev2.Event)(nil), &store.ErrNotFound{})
	ms.On("UpdateEvent", mock.Anything, mock.Anything).Return((*corev2.Event)(nil), (*corev2.Event)(nil), nil)
	if _, _, err := s.UpdateEvent(ctx, event); err != nil {
		t.Fatal(err)
	}
	if got := readEvent(s.db, "default", "entity1", "check1"); got.HasMetrics() {
		t.Error("event has metrics but should not")
	}
}

func TestUpdateEventHasCheckState(t *testing.T) {
	ms := new(mockstore.MockStore)
	config := EventStoreConfig{
		BackingStore:    ms,
		FlushInterval:   10 * time.Second,
		SilenceStore:    new(mockstore.MockStore),
		EventWriteLimit: 1000,
	}
	s := NewEventStore(config)
	event := fixtureEvent("foo", "bar")
	ctx := context.Background()
	ms.On("GetEventByEntityCheck", mock.Anything, mock.Anything, mock.Anything).Return((*corev2.Event)(nil), &store.ErrNotFound{})
	ms.On("UpdateEvent", mock.Anything, mock.Anything).Return((*corev2.Event)(nil), (*corev2.Event)(nil), nil)
	updatedEvent, previousEvent, err := s.UpdateEvent(ctx, event)
	if err != nil {
		t.Fatal(err)
	}
	if previousEvent != nil {
		t.Errorf("previous event is not nil")
	}
	if got, want := updatedEvent.Check.State, corev2.EventPassingState; got != want {
		t.Fatalf("bad check state: got %q, want %q", got, want)
	}
}
