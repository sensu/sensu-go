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
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
)

// EntityClient is an API client for entities.
type EntityClient struct {
	client     *GenericClient
	storev2    storev2.Interface
	eventStore store.EventStore
	auth       authorization.Authorizer
}

// NewEntityClient creates a new EntityClient given a store, an event store and
// an authorizer.
func NewEntityClient(store store.ResourceStore, storev2 storev2.Interface, eventStore store.EventStore, auth authorization.Authorizer) *EntityClient {
	return &EntityClient{
		client: &GenericClient{
			Auth:       auth,
			Store:      store,
			Kind:       &corev2.Entity{},
			APIGroup:   "core",
			APIVersion: "v2",
		},
		storev2:    storev2,
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
		return err
	}
	return nil
}

// UpdateEntity updates an entity, if authorized.
func (e *EntityClient) UpdateEntity(ctx context.Context, entity *corev2.Entity) error {
	// TODO(ccressent): add blurb here to explain why we make that exception.
	if entity.EntityClass == corev2.EntityProxyClass {
		if err := e.client.Update(ctx, entity); err != nil {
			return err
		}
	} else {
		// The generic client takes care of authorization for us, so if we
		// bypass it as we're doing here, we must not forget to deal with
		// authorization ourselves.
		attrs := entityUpdateAttributes(ctx, entity.Name)
		if err := authorize(ctx, e.auth, attrs); err != nil {
			return err
		}

		config, _ := corev3.V2EntityToV3(entity)
		req := storev2.NewResourceRequestFromResource(ctx, config)

		wConfig, err := wrap.Resource(config)
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
	var entity corev2.Entity
	if err := e.client.Get(ctx, name, &entity); err != nil {
		return nil, err
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
		return nil, err
	}
	return slice, nil
}

func entityUpdateAttributes(ctx context.Context, name string) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:     "core",
		APIVersion:   "v2",
		Namespace:    corev2.ContextNamespace(ctx),
		Resource:     "entities",
		Verb:         "update",
		ResourceName: name,
	}
}
