package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// EventController expose actions in which a viewer can perform.
type EventController struct {
	Store store.EventStore
	Bus   messaging.MessageBus
}

// NewEventController returns new EventController
func NewEventController(store store.EventStore, bus messaging.MessageBus) EventController {
	return EventController{
		Store: store,
		Bus:   bus,
	}
}

// Query returns resources available to the viewer filter by given params.
func (a EventController) Query(ctx context.Context, entityID, checkName string) ([]*types.Event, error) {
	var results []*types.Event

	// Fetch from store
	var serr error
	if entityID != "" && checkName != "" {
		var result *types.Event
		result, serr = a.Store.GetEventByEntityCheck(ctx, entityID, checkName)
		if result != nil {
			results = append(results, result)
		}
	} else if entityID != "" {
		results, serr = a.Store.GetEventsByEntity(ctx, entityID)
	} else {
		results, serr = a.Store.GetEvents(ctx)
	}

	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	return results, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (a EventController) Find(ctx context.Context, entity, check string) (*types.Event, error) {
	// Find (for events) requires both an entity and check
	if entity == "" || check == "" {
		return nil, NewErrorf(InvalidArgument, "Find() requires both an entity and a check")
	}

	result, err := a.Store.GetEventByEntityCheck(ctx, entity, check)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	return result, nil
}

// Destroy destroys the event indicated by the supplied entity and check.
func (a EventController) Destroy(ctx context.Context, entity, check string) error {
	// Destroy (for events) requires both an entity and check
	if entity == "" || check == "" {
		return NewErrorf(InvalidArgument, "Destroy() requires both an entity and a check")
	}

	result, err := a.Store.GetEventByEntityCheck(ctx, entity, check)
	if err != nil {
		return NewError(InternalErr, err)
	}

	if result != nil {
		err := a.Store.DeleteEventByEntityCheck(ctx, entity, check)
		if err != nil {
			return NewError(InternalErr, err)
		}
	}

	return nil
}

// Create creates the event indicated by the supplied entity and check.
// If an event already exists for the entity and check, it updates that event.
func (a EventController) Create(ctx context.Context, event types.Event) error {
	if err := event.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Verify if we already have an existing event for this entity/check pair.
	// Doesn't apply to metric events.
	if event.HasCheck() {
		check := event.Check
		entity := event.Entity

		e, err := a.Store.GetEventByEntityCheck(ctx, entity.ID, check.Name)
		if err != nil {
			return NewError(InternalErr, err)
		} else if e != nil {
			return NewErrorf(AlreadyExistsErr)
		}
	}

	// Publish to event pipeline
	if err := a.Bus.Publish(messaging.TopicEventRaw, &event); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// CreateOrReplace creates the event indicated by the supplied entity and check.
// If an event already exists for the entity and check, it updates that event.
func (a EventController) CreateOrReplace(ctx context.Context, event types.Event) error {
	if err := event.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Publish to event pipeline
	if err := a.Bus.Publish(messaging.TopicEventRaw, &event); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}
