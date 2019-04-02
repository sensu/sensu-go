package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

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

// Find returns resource associated with given parameters if available to the
// viewer.
func (a EventController) Find(ctx context.Context, entity, check string) (*corev2.Event, error) {
	// Find (for events) requires both an entity and check
	if entity == "" || check == "" {
		return nil, NewErrorf(InvalidArgument, "Find() requires both an entity and a check")
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

// Destroy destroys the event indicated by the supplied entity and check.
func (a EventController) Destroy(ctx context.Context, entity, check string) error {
	// Destroy (for events) requires both an entity and check
	if entity == "" || check == "" {
		return NewErrorf(InvalidArgument, "Destroy() requires both an entity and a check")
	}

	result, err := a.store.GetEventByEntityCheck(ctx, entity, check)
	if err != nil {
		return NewError(InternalErr, err)
	}

	if result != nil {
		err := a.store.DeleteEventByEntityCheck(ctx, entity, check)
		if err != nil {
			return NewError(InternalErr, err)
		}
	}

	return nil
}

// Create creates the event indicated by the supplied entity and check.
// If an event already exists for the entity and check, it updates that event.
func (a EventController) Create(ctx context.Context, event corev2.Event) error {
	if err := event.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Verify if we already have an existing event for this entity/check pair.
	// Doesn't apply to metric events.
	if event.HasCheck() {
		check := event.Check
		entity := event.Entity

		e, err := a.store.GetEventByEntityCheck(ctx, entity.Name, check.Name)
		if err != nil {
			return NewError(InternalErr, err)
		} else if e != nil {
			return NewErrorf(AlreadyExistsErr)
		}
	}

	// Publish to event pipeline
	if err := a.bus.Publish(messaging.TopicEventRaw, &event); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// CreateOrReplace creates the event indicated by the supplied entity and check.
// If an event already exists for the entity and check, it updates that event.
func (a EventController) CreateOrReplace(ctx context.Context, event corev2.Event) error {
	if err := event.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Publish to event pipeline
	if err := a.bus.Publish(messaging.TopicEventRaw, &event); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}
