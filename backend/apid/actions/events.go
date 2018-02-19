package actions

import (
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

// eventUpdateFields whitelists fields allowed to be updated for Events
var eventUpdateFields = []string{
	"Timestamp",
	"Entity",
	"Check",
}

// EventController expose actions in which a viewer can perform.
type EventController struct {
	Store  store.EventStore
	Policy authorization.EventPolicy
	Bus    messaging.MessageBus
}

// NewEventController returns new EventController
func NewEventController(store store.EventStore, bus messaging.MessageBus) EventController {
	return EventController{
		Store:  store,
		Policy: authorization.Events,
		Bus:    bus,
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

	// Filter out those resources the viewer does not have access to view.
	abilities := a.Policy.WithContext(ctx)
	for i := 0; i < len(results); i++ {
		if !abilities.CanRead(results[i]) {
			results = append(results[:i], results[i+1:]...)
			i--
		}
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

	// Verify user has permission to view
	abilities := a.Policy.WithContext(ctx)
	if result != nil && abilities.CanRead(result) {
		return result, nil
	}

	return nil, NewErrorf(NotFound)
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

	// Verify user has permission to delete
	abilities := a.Policy.WithContext(ctx)
	if result != nil && abilities.CanDelete() {
		err := a.Store.DeleteEventByEntityCheck(ctx, entity, check)
		if err != nil {
			err = NewError(InternalErr, err)
		}
		return err
	}

	return NewErrorf(NotFound)
}

// Update updates the event indicated by the supplied entity and check.
func (a EventController) Update(ctx context.Context, event types.Event) error {
	check := event.Check
	entity := event.Entity

	// Adjust context
	policy := a.Policy.WithContext(ctx)

	// Validate
	if err := event.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Check for existing
	e, err := a.Store.GetEventByEntityCheck(ctx, entity.ID, check.Name)
	if err != nil {
		return NewError(InternalErr, err)
	} else if e == nil {
		return NewErrorf(NotFound)
	}

	// Copy
	copyFields(e, &event, eventUpdateFields...)

	// Verify viewer can make change
	if ok := policy.CanUpdate(e); !ok {
		return NewErrorf(PermissionDenied, "update")
	}

	// Publish to event pipeline
	if err := a.Bus.Publish(messaging.TopicEventRaw, &event); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Create creates the event indicated by the supplied entity and check.
// If an event already exists for the entity and check, it updates that event.
func (a EventController) Create(ctx context.Context, event types.Event) error {
	check := event.Check
	entity := event.Entity

	// Adjust context
	policy := a.Policy.WithContext(ctx)

	// Validate
	if err := event.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Check for existing
	e, err := a.Store.GetEventByEntityCheck(ctx, entity.ID, check.Name)
	if err != nil {
		return NewError(InternalErr, err)
	} else if e != nil {
		return a.Update(ctx, event)
	}

	// Verify permissions
	if ok := policy.CanCreate(&event); !ok {
		return NewErrorf(PermissionDenied, "create")
	}

	// Publish to event pipeline
	if err := a.Bus.Publish(messaging.TopicEventRaw, &event); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}
