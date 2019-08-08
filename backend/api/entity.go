package api

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sirupsen/logrus"
)

// EntityClient is an API client for entities.
type EntityClient struct {
	client     *GenericClient
	eventStore store.EventStore
	auth       authorization.Authorizer
}

// NewEntityClient creates a new EntityClient given a store, an event store and
// an authorizer.
func NewEntityClient(store store.ResourceStore, eventStore store.EventStore, auth authorization.Authorizer) *EntityClient {
	return &EntityClient{
		client: &GenericClient{
			Kind:       &corev2.Entity{},
			Auth:       auth,
			Store:      store,
			APIGroup:   "core",
			APIVersion: "v2",
		},
		eventStore: eventStore,
		auth:       auth,
	}
}

// DeleteEntity deletes an Entity, if authorized. In doing so, it will also
// delete all of the events associated with the entity. The operation is not
// transactional; partial data may remain if it fails.
func (e *EntityClient) DeleteEntity(ctx context.Context, name string) error {
	if err := e.client.Delete(ctx, name); err != nil {
		return err
	}
	// if the client delete succeeded, we have sufficient permissions to also
	// delete the associated events.
	events, err := e.eventStore.GetEventsByEntity(ctx, name, &store.SelectionPredicate{})
	if err != nil {
		return fmt.Errorf("could not delete events associated with entity: %s", err)
	}
	for _, event := range events {
		if !event.HasCheck() {
			// improbable
			continue
		}
		if err := e.eventStore.DeleteEventByEntityCheck(ctx, name, event.Check.Name); err != nil {
			logger := logger.WithFields(logrus.Fields{
				"entity":    name,
				"check":     event.Check.Name,
				"namespace": event.Namespace})
			logger.WithError(err).Error("error deleting event from entity")
			continue
		}
	}

	return nil
}

// CreateEntity creates an entity, if authorized.
func (e *EntityClient) CreateEntity(ctx context.Context, entity *corev2.Entity) error {
	if err := e.client.Create(ctx, entity); err != nil {
		return fmt.Errorf("couldn't create entity: %s", err)
	}
	return nil
}

// UpdateEntity updates an entity, if authorized.
func (e *EntityClient) UpdateEntity(ctx context.Context, entity *corev2.Entity) error {
	if err := e.client.Update(ctx, entity); err != nil {
		return fmt.Errorf("couldn't update entity: %s", err)
	}
	return nil
}

// FetchEntity gets an entity, if authorized.
func (e *EntityClient) FetchEntity(ctx context.Context, name string) (*corev2.Entity, error) {
	var entity corev2.Entity
	if err := e.client.Get(ctx, name, &entity); err != nil {
		return nil, fmt.Errorf("couldn't get entity: %s", err)
	}
	return &entity, nil
}

// ListEntities lists all entities in a namespace, if authorized.
func (e *EntityClient) ListEntities(ctx context.Context) ([]*corev2.Entity, error) {
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
	slice := []*corev2.Entity{}
	if err := e.client.List(ctx, &slice, pred); err != nil {
		return nil, fmt.Errorf("couldn't list entities: %s", err)
	}
	return slice, nil
}
