package api

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
)

type Publisher interface {
	Publish(topic string, message interface{}) error
}

type EventClient struct {
	store store.EventStore
	auth  authorization.Authorizer
	bus   Publisher
}

func NewEventClient(store store.EventStore, auth authorization.Authorizer, bus Publisher) *EventClient {
	return &EventClient{
		store: store,
		auth:  auth,
		bus:   bus,
	}
}

func (e *EventClient) UpdateEvent(ctx context.Context, event *corev2.Event) error {
	if err := event.Validate(); err != nil {
		return fmt.Errorf("couldn't create event: %s", err)
	}
	attrs := eventCreateAttributes(ctx)
	if err := authorize(ctx, e.auth, attrs); err != nil {
		return fmt.Errorf("couldn't create event: %s", err)
	}
	// Update the event through eventd
	return e.bus.Publish(messaging.TopicEventRaw, event)
}

func (e *EventClient) GetEvent(ctx context.Context, entity, check string) (*corev2.Event, error) {
	attrs := eventGetAttributes(ctx, fmt.Sprintf("%s:%s", entity, check))
	if err := authorize(ctx, e.auth, attrs); err != nil {
		return nil, fmt.Errorf("couldn't get event: %s", err)
	}
	return e.store.GetEventByEntityCheck(ctx, entity, check)
}

func (e *EventClient) DeleteEvent(ctx context.Context, entity, check string) error {
	attrs := eventDeleteAttributes(ctx, entity, check)
	if err := authorize(ctx, e.auth, attrs); err != nil {
		return fmt.Errorf("couldn't delete event: %s", err)
	}
	if err := e.store.DeleteEventByEntityCheck(ctx, entity, check); err != nil {
		return fmt.Errorf("couldn't delete event: %s", err)
	}
	return nil
}

func (e *EventClient) ListEvents(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
	attrs := eventListAttributes(ctx)
	if err := authorize(ctx, e.auth, attrs); err != nil {
		return nil, fmt.Errorf("couldn't list events: %s", err)
	}
	events, err := e.store.GetEvents(ctx, pred)
	if err != nil {
		return nil, fmt.Errorf("couldn't list events: %s", err)
	}
	return events, nil
}

func eventCreateAttributes(ctx context.Context) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:   "core",
		APIVersion: "v2",
		Namespace:  corev2.ContextNamespace(ctx),
		Resource:   "events",
		Verb:       "create,update",
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
		Verb:         "del",
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
