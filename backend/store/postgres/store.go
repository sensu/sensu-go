package postgres

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// Store is like sensu-go's store, but part of it is backed by postgresql.
type Store struct {
	store.Store
	store.EventStore
	store.EntityStore
	store.NamespaceStore
	store.SilenceStore
}

// DeleteSilences deletes all silences matching the given names.
func (s Store) DeleteSilences(ctx context.Context, namespace string, names []string) error {
	return s.SilenceStore.DeleteSilences(ctx, namespace, names)
}

// GetSilences returns all silences in the namespace. A nil slice with no error is
// returned if none were found.
func (s Store) GetSilences(ctx context.Context, namespace string) ([]*corev2.Silenced, error) {
	return s.SilenceStore.GetSilences(ctx, namespace)
}

// GetSilencedsByCheck returns all silences for the given check . A nil
// slice with no error is returned if none were found.
func (s Store) GetSilencesByCheck(ctx context.Context, namespace, check string) ([]*corev2.Silenced, error) {
	return s.SilenceStore.GetSilencesByCheck(ctx, namespace, check)
}

// GetSilencedEntriesBySubscription returns all entries for the given
// subscription. A nil slice with no error is returned if none were found.
func (s Store) GetSilencesBySubscription(ctx context.Context, namespace string, subscriptions []string) ([]*corev2.Silenced, error) {
	return s.SilenceStore.GetSilencesBySubscription(ctx, namespace, subscriptions)
}

// GetSilenceByName returns an entry using the given namespace and name. An
// error is returned if the entry is not found.
func (s Store) GetSilenceByName(ctx context.Context, namespace, name string) (*corev2.Silenced, error) {
	return s.SilenceStore.GetSilenceByName(ctx, namespace, name)
}

// UpdateSilence creates or updates a given silence.
func (s Store) UpdateSilence(ctx context.Context, silence *corev2.Silenced) error {
	return s.SilenceStore.UpdateSilence(ctx, silence)
}

// GetSilencesByName gets all the named silence entries.
func (s Store) GetSilencesByName(ctx context.Context, namespace string, names []string) ([]*corev2.Silenced, error) {
	return s.SilenceStore.GetSilencesByName(ctx, namespace, names)
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

func (s Store) CreateNamespace(ctx context.Context, namespace *corev2.Namespace) error {
	return s.NamespaceStore.CreateNamespace(ctx, namespace)
}

func (s Store) DeleteNamespace(ctx context.Context, name string) error {
	return s.NamespaceStore.DeleteNamespace(ctx, name)
}

func (s Store) GetNamespace(ctx context.Context, name string) (*corev2.Namespace, error) {
	return s.NamespaceStore.GetNamespace(ctx, name)
}

func (s Store) ListNamespaces(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Namespace, error) {
	return s.NamespaceStore.ListNamespaces(ctx, pred)
}

func (s Store) UpdateNamespace(ctx context.Context, namespace *corev2.Namespace) error {
	return s.NamespaceStore.UpdateNamespace(ctx, namespace)
}
