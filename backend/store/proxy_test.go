package store

import (
	"context"
	"errors"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

type mockEventStore struct {
	id string
}

func (m mockEventStore) DeleteEventByEntityCheck(ctx context.Context, entity, check string) error {
	return errors.New(m.id)
}

func (mockEventStore) GetEvents(ctx context.Context, pred *SelectionPredicate) ([]*corev2.Event, error) {
	return nil, nil
}

func (mockEventStore) GetEventsByEntity(ctx context.Context, entity string, pred *SelectionPredicate) ([]*corev2.Event, error) {
	return nil, nil
}

func (mockEventStore) GetEventByEntityCheck(ctx context.Context, entity, check string) (*types.Event, error) {
	return nil, nil
}

func (mockEventStore) UpdateEvent(ctx context.Context, event *types.Event) (old, new *types.Event, err error) {
	return nil, nil, nil
}

func TestEventStoreProxy(t *testing.T) {
	storeA := mockEventStore{"a"}
	storeB := mockEventStore{"b"}

	proxy := NewEventStoreProxy(storeA)
	ctx := context.Background()

	err := proxy.DeleteEventByEntityCheck(ctx, "", "")
	if got, want := err.Error(), "a"; got != want {
		t.Fatalf("bad response from event store: got %s, want %s", got, want)
	}

	proxy.UpdateEventStore(storeB)

	err = proxy.DeleteEventByEntityCheck(ctx, "", "")
	if got, want := err.Error(), "b"; got != want {
		t.Fatalf("bad response from event store: got %s, want %s", got, want)
	}
}
