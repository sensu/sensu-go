package api

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// EntityClient is an API client for entities.
type EntityClient struct {
	storev2     storev2.Interface
	entityStore store.EntityStore
	eventStore  store.EventStore
	auth        authorization.Authorizer
}

// NewEntityClient creates a new EntityClient given a store, an event store and
// an authorizer.
func NewEntityClient(store store.EntityStore, storev2 storev2.Interface, eventStore store.EventStore, auth authorization.Authorizer) *EntityClient {
	return &EntityClient{
		storev2:     storev2,
		entityStore: store,
		eventStore:  eventStore,
		auth:        auth,
	}
}

// DeleteEntity deletes an Entity, if authorized. In doing so, it will also
// delete all of the events associated with the entity. The operation is not
// transactional; partial data may remain if it fails.
func (e *EntityClient) DeleteEntity(ctx context.Context, name string) error {
	attrs := entityAuthAttributes(ctx, "delete", name)
	if err := authorize(ctx, e.auth, attrs); err != nil {
		return err
	}
	if err := e.entityStore.DeleteEntityByName(ctx, name); err != nil {
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
	attrs := entityAuthAttributes(ctx, "create", entity.Name)
	if err := authorize(ctx, e.auth, attrs); err != nil {
		return err
	}
	setCreatedBy(ctx, entity)
	if err := e.entityStore.UpdateEntity(ctx, entity); err != nil {
		return err
	}
	return nil
}

// UpdateEntity updates an entity, if authorized.
func (e *EntityClient) UpdateEntity(ctx context.Context, entity *corev2.Entity) error {
	attrs := entityAuthAttributes(ctx, "update", entity.Name)
	if err := authorize(ctx, e.auth, attrs); err != nil {
		return err
	}
	setCreatedBy(ctx, entity)

	// We have 2 code paths here: one for proxy entities and another for all
	// other types of entities. We had to make that distinction because Entity
	// is still the public API to interact with entities, even though internally
	// we use the storev2 EntityConfig/EntityState split.
	//
	// The consequence was that updating an Entity could alter its state,
	// something we don't really want unless that entity is a proxy entity.
	//
	// See sensu-go#3896.
	if entity.EntityClass == corev2.EntityProxyClass {
		if err := e.entityStore.UpdateEntity(ctx, entity); err != nil {
			return err
		}
	} else {
		config, _ := corev3.V2EntityToV3(entity)
		// Ensure per-entity subscription does not get removed
		config.Subscriptions = corev2.AddEntitySubscription(config.Metadata.Name, config.Subscriptions)
		req := storev2.NewResourceRequestFromResource(ctx, config)

		wConfig, err := storev2.WrapResource(config)
		if err != nil {
			return err
		}

		if err := e.storev2.CreateOrUpdate(req, wConfig); err != nil {
			return err
		}
	}

	return nil
}

// FetchEntity gets an entity, if authorized.
func (e *EntityClient) FetchEntity(ctx context.Context, name string) (*corev2.Entity, error) {
	attrs := entityAuthAttributes(ctx, "get", name)
	if err := authorize(ctx, e.auth, attrs); err != nil {
		return nil, err
	}
	entity, err := e.entityStore.GetEntityByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// ListEntities lists all entities in a namespace, if authorized.
func (e *EntityClient) ListEntities(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Entity, error) {
	attrs := entityAuthAttributes(ctx, "list", "")
	if err := authorize(ctx, e.auth, attrs); err != nil {
		return nil, err
	}
	slice, err := e.entityStore.GetEntities(ctx, pred)
	if err != nil {
		return nil, err
	}
	return slice, nil
}

func entityAuthAttributes(ctx context.Context, verb, name string) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:     "core",
		APIVersion:   "v2",
		Namespace:    corev2.ContextNamespace(ctx),
		Resource:     "entities",
		Verb:         verb,
		ResourceName: name,
	}
}
