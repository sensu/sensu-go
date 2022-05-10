package postgres

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// Store is like sensu-go's store, but part of it is backed by postgresql.
type Store struct {
	store.Store
	store.EventStore
	store.EntityStore
	Config backend.PostgresConfig
}

func (s Store) DeleteEventByEntityCheck(ctx context.Context, entity, check string) error {
	return s.EventStore.DeleteEventByEntityCheck(ctx, entity, check)
}

func (s Store) GetEvents(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
	return s.EventStore.GetEvents(ctx, pred)
}

func (s Store) GetEventsByEntity(ctx context.Context, entity string, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
	return s.EventStore.GetEventsByEntity(ctx, entity, pred)
}

func (s Store) GetEventByEntityCheck(ctx context.Context, entity, check string) (*corev2.Event, error) {
	return s.EventStore.GetEventByEntityCheck(ctx, entity, check)
}

func (s Store) UpdateEvent(ctx context.Context, event *corev2.Event) (old, new *corev2.Event, err error) {
	return s.EventStore.UpdateEvent(ctx, event)
}

func (s Store) DeleteEntity(ctx context.Context, entity *types.Entity) error {
	return s.EntityStore.DeleteEntity(ctx, entity)
}

func (s Store) DeleteEntityByName(ctx context.Context, name string) error {
	return s.EntityStore.DeleteEntityByName(ctx, name)
}

func (s Store) GetEntities(ctx context.Context, pred *store.SelectionPredicate) ([]*types.Entity, error) {
	return s.EntityStore.GetEntities(ctx, pred)
}

func (s Store) GetEntityByName(ctx context.Context, name string) (*types.Entity, error) {
	return s.EntityStore.GetEntityByName(ctx, name)
}

func (s Store) UpdateEntity(ctx context.Context, entity *types.Entity) error {
	return s.EntityStore.UpdateEntity(ctx, entity)
}

func (s Store) CountEvents(ctx context.Context, pred *store.SelectionPredicate) (int64, error) {
	return s.EventStore.CountEvents(ctx, pred)
}

func (s Store) EventStoreSupportsFiltering(ctx context.Context) bool {
	return s.EventStore.EventStoreSupportsFiltering(ctx)
}
