package actions

import (
	"context"

	"github.com/google/uuid"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

const deletedEventSentinel = -1

// EventController expose actions in which a viewer can perform.
type EventController struct {
	store store.EventStore
	bus   messaging.MessageBus
}

// NewEventController returns new EventController
func NewEventController(store store.EventStore, bus messaging.MessageBus) EventController {
	return EventController{
		store: store,
		bus:   bus,
	}
}

// List returns resources available to the viewer filter by given params.
func (a EventController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	var results []*corev2.Event
	var err error

	// Fetch from store
	if pred.Subcollection != "" {
		results, err = a.store.GetEventsByEntity(ctx, pred.Subcollection, pred)
	} else {
		results, err = a.store.GetEvents(ctx, pred)
	}

	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	resources := make([]corev2.Resource, len(results))
	for i, v := range results {
		resources[i] = corev2.Resource(v)
	}

	return resources, nil
}

// Get returns resource associated with given parameters if available to the
// viewer.
func (a EventController) Get(ctx context.Context, entity, check string) (*corev2.Event, error) {
	// Get (for events) requires both an entity and check
	if entity == "" || check == "" {
		return nil, NewErrorf(InvalidArgument, "Get() requires both an entity and a check")
	}

	result, err := a.store.GetEventByEntityCheck(ctx, entity, check)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}
	if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}

// Delete destroys the event indicated by the supplied entity and check.
func (a EventController) Delete(ctx context.Context, entity, check string) error {
	// Destroy (for events) requires both an entity and check
	if entity == "" || check == "" {
		return NewErrorf(InvalidArgument, "Delete() requires both an entity and a check")
	}

	result, err := a.store.GetEventByEntityCheck(ctx, entity, check)
	if err != nil {
		return NewError(InternalErr, err)
	}

	if result == nil {
		return NewErrorf(NotFound)
	}

	if result.HasCheck() && result.Check.Ttl > 0 {
		// Disable check TTL for this event, and inform eventd
		result.Check.Ttl = deletedEventSentinel
		if err := a.bus.Publish(messaging.TopicEventRaw, result); err != nil {
			return NewError(InternalErr, err)
		}
	}

	if result.HasCheck() && result.Check.Name == "keepalive" {
		// Notify keepalived that the keepalive was deleted
		result.Timestamp = deletedEventSentinel
		if err := a.bus.Publish(messaging.TopicKeepalive, result); err != nil {
			return NewError(InternalErr, err)
		}
	}

	if err := a.store.DeleteEventByEntityCheck(ctx, entity, check); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// CreateOrReplace creates the event indicated by the supplied entity and check.
// If an event already exists for the entity and check, it updates that event.
func (a EventController) CreateOrReplace(ctx context.Context, event *corev2.Event) error {
	if event.Entity != nil && event.Entity.EntityClass == "" {
		event.Entity.EntityClass = corev2.EntityProxyClass
	}

	if err := event.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	if len(event.ID) == 0 {
		id, err := uuid.NewRandom()
		if err != nil {
			return NewError(InternalErr, err)
		}
		event.ID = id[:]
	}

	if claims := jwt.GetClaimsFromContext(ctx); claims != nil {
		event.CreatedBy = claims.StandardClaims.Subject
		event.Check.CreatedBy = claims.StandardClaims.Subject
		event.Entity.CreatedBy = claims.StandardClaims.Subject
	}

	// Publish to event pipeline
	if err := a.bus.Publish(messaging.TopicEventRaw, event); err != nil {
		return NewError(InternalErr, err)
	}

	if event.HasCheck() && event.Check.Name == "keepalive" {
		if err := a.bus.Publish(messaging.TopicKeepalive, event); err != nil {
			return NewError(InternalErr, err)
		}
	}

	return nil
}
