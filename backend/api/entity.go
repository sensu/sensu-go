package api

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sirupsen/logrus"
)

type EntityClient struct {
	client     *genericClient
	eventStore store.EventStore
	auth       authorization.Authorizer
}

func NewEntityClient(store store.ResourceStore, eventStore store.EventStore, auth authorization.Authorizer) *EntityClient {
	return &EntityClient{
		client: &genericClient{
			Kind:       &corev2.Entity{},
			Auth:       auth,
			Resource:   "entities",
			APIGroup:   "core",
			APIVersion: "v2",
			Store:      store,
		},
		eventStore: eventStore,
		auth:       auth,
	}
}

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

func (e *EntityClient) CreateEntity(ctx context.Context, entity *corev2.Entity) error {
	if err := e.client.Create(ctx, entity); err != nil {
		return fmt.Errorf("couldn't create entity: %s", err)
	}
	return nil
}

func (e *EntityClient) UpdateEntity(ctx context.Context, entity *corev2.Entity) error {
	if err := e.client.Update(ctx, entity); err != nil {
		return fmt.Errorf("couldn't update entity: %s", err)
	}
	return nil
}

func (e *EntityClient) GetEntity(ctx context.Context, name string) (*corev2.Entity, error) {
	var entity corev2.Entity
	if err := e.client.Get(ctx, name, &entity); err != nil {
		return nil, fmt.Errorf("couldn't get entity: %s", err)
	}
	return &entity, nil
}
