package api

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
)

// Publisher is an interface that represents the message bus concept.
type Publisher interface {
	Publish(topic string, message interface{}) error
}

// EventClient is an API client for events.
type EventClient struct {
	store store.EventStore
	auth  authorization.Authorizer
	bus   Publisher
}

// EventStoreSupportsFiltering stub impl
func (c EventClient) EventStoreSupportsFiltering(ctx context.Context) bool {
	return c.store.EventStoreSupportsFiltering(ctx)
}

// CountEvents proxies to store client
func (c EventClient) CountEvents(ctx context.Context, pred *store.SelectionPredicate) (int64, error) {
	return c.store.CountEvents(ctx, pred)
}

// NewEventClient creates a new EventClient, given a store, authorizer, and bus.
func NewEventClient(store store.EventStore, auth authorization.Authorizer, bus Publisher) *EventClient {
	return &EventClient{
		store: store,
		auth:  auth,
		bus:   bus,
	}
}

// UpdateEvent updates an event, and publishes the update ot the bus, if
// authorized.
func (e *EventClient) UpdateEvent(ctx context.Context, event *corev2.Event) error {
	if err := event.Validate(); err != nil {
		return fmt.Errorf("couldn't create event: %s", err)
	}
	attrs := eventUpdateAttributes(ctx)
	if err := authorize(ctx, e.auth, attrs); err != nil {
		return err
	}
	if claims := jwt.GetClaimsFromContext(ctx); claims != nil {
		event.CreatedBy = claims.StandardClaims.Subject
		event.Check.CreatedBy = claims.StandardClaims.Subject
		event.Entity.CreatedBy = claims.StandardClaims.Subject
	}
	// Update the event through eventd
	return e.bus.Publish(messaging.TopicEventRaw, event)
}

// FetchEvent gets an event, if authorized.
func (e *EventClient) FetchEvent(ctx context.Context, entity, check string) (*corev2.Event, error) {
	attrs := eventGetAttributes(ctx, fmt.Sprintf("%s:%s", entity, check))
	if err := authorize(ctx, e.auth, attrs); err != nil {
		return nil, err
	}
	return e.store.GetEventByEntityCheck(ctx, entity, check)
}

// DeleteEvent deletes an event, if authorized.
func (e *EventClient) DeleteEvent(ctx context.Context, entity, check string) error {
	attrs := eventDeleteAttributes(ctx, entity, check)
	if err := authorize(ctx, e.auth, attrs); err != nil {
		return err
	}
	if err := e.store.DeleteEventByEntityCheck(ctx, entity, check); err != nil {
		return fmt.Errorf("couldn't delete event: %s", err)
	}
	return nil
}

// ListEvents lists all events in a namespace, according to the selection
// predicate, if authorized.
func (e *EventClient) ListEvents(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
	attrs := eventListAttributes(ctx)
	if err := authorize(ctx, e.auth, attrs); err != nil {
		return nil, err
	}
	events, err := e.store.GetEvents(ctx, pred)
	if err != nil {
		return nil, fmt.Errorf("couldn't list events: %s", err)
	}
	return events, nil
}

// ListEventsByEntity lists all events in a namespace, according to the
// selection predicate, if authorized.
func (e *EventClient) ListEventsByEntity(ctx context.Context, entity string, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
	attrs := eventListAttributes(ctx)
	if err := authorize(ctx, e.auth, attrs); err != nil {
		return nil, err
	}
	events, err := e.store.GetEventsByEntity(ctx, entity, pred)
	if err != nil {
		return nil, fmt.Errorf("couldn't list events by entity: %s", err)
	}
	return events, nil
}

func eventUpdateAttributes(ctx context.Context) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:   "core",
		APIVersion: "v2",
		Namespace:  corev2.ContextNamespace(ctx),
		Resource:   "events",
		Verb:       "update",
	}
}

func eventGetAttributes(ctx context.Context, name string) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:     "core",
		APIVersion:   "v2",
		Namespace:    corev2.ContextNamespace(ctx),
		Resource:     "events",
		Verb:         "get",
		ResourceName: name,
	}
}

func eventDeleteAttributes(ctx context.Context, entity, check string) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:     "core",
		APIVersion:   "v2",
		Namespace:    corev2.ContextNamespace(ctx),
		Resource:     "events",
		Verb:         "delete",
		ResourceName: fmt.Sprintf("%s:%s", entity, check),
	}
}

func eventListAttributes(ctx context.Context) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:   "core",
		APIVersion: "v2",
		Namespace:  corev2.ContextNamespace(ctx),
		Resource:   "events",
		Verb:       "list",
	}
}
