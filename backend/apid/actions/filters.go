package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var filterUpdateFields = []string{
	"Action",
	"When",
	"Expressions",
	"RuntimeAssets",
}

// EventFilterController allows querying EventFilters in bulk or by name.
type EventFilterController struct {
	Store store.EventFilterStore
}

// NewEventFilterController creates a new EventFilterController backed by store.
func NewEventFilterController(store store.EventFilterStore) EventFilterController {
	return EventFilterController{
		Store: store,
	}
}

// Create creates a new EventFilter resource.
// It returns non-nil error if the new Filter is invalid, update permissions
// do not exist, or an internal error occurs while updating the underlying
// store.
func (c EventFilterController) Create(ctx context.Context, filter types.EventFilter) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &filter)

	// Check for existing
	if m, err := c.Store.GetEventFilterByName(ctx, filter.Name); err != nil {
		return NewError(InternalErr, err)
	} else if m != nil {
		return NewErrorf(AlreadyExistsErr, filter.Name)
	}

	// Validate
	if err := filter.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := c.Store.UpdateEventFilter(ctx, &filter); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// CreateOrReplace creates a new EventFilter resource.
// It returns non-nil error if the new Filter is invalid, update permissions
// do not exist, or an internal error occurs while updating the underlying
// store.
func (c EventFilterController) CreateOrReplace(ctx context.Context, filter types.EventFilter) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &filter)

	// Validate
	if err := filter.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := c.Store.UpdateEventFilter(ctx, &filter); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Update updates a Filter.
// It returns non-nil error if the new Filter is invalid, create permissions
// do not exist, or an internal error occurs while updating the underlying
// store.
func (c EventFilterController) Update(ctx context.Context, delta types.EventFilter) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &delta)

	// Check for existing
	filter, err := c.Store.GetEventFilterByName(ctx, delta.Name)
	if err != nil {
		return NewError(InternalErr, err)
	} else if filter == nil {
		return NewErrorf(NotFound, delta.Name)
	}

	// Update
	if err := filter.Update(&delta, filterUpdateFields...); err != nil {
		return NewError(InternalErr, err)
	}

	// Validate
	if err := filter.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := c.Store.UpdateEventFilter(ctx, filter); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Query returns resources available to the viewer filter by given params.
// It returns non-nil error if the params are invalid, read permissions
// do not exist, or an internal error occurs while reading the underlying
// store.
func (c EventFilterController) Query(ctx context.Context) ([]*types.EventFilter, error) {

	// Fetch from store
	filters, err := c.Store.GetEventFilters(ctx)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	return filters, nil
}

// Destroy destroys the named EventFilter.
// It returns non-nil error if the params are invalid, delete permissions
// do not exist, or an internal error occurs while updating the underlying
// Store.
func (c EventFilterController) Destroy(ctx context.Context, name string) error {
	// Validate parameters
	if name == "" {
		return NewErrorf(InvalidArgument, "name is undefined")
	}

	// Fetch from store
	filter, err := c.Store.GetEventFilterByName(ctx, name)
	if err != nil {
		return NewError(InternalErr, err)
	}
	if filter == nil {
		return NewErrorf(NotFound, name)
	}

	// Remove from store
	if err := c.Store.DeleteEventFilterByName(ctx, filter.Name); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
// It returns non-nil error if the params are invalid, read permissions
// do not exist, or an internal error occurs while reading the underlying
// Store.
func (c EventFilterController) Find(ctx context.Context, name string) (*types.EventFilter, error) {
	// Find (for mutators) requires a name
	if name == "" {
		return nil, NewErrorf(InvalidArgument, "Find() requires a name")
	}

	result, err := c.Store.GetEventFilterByName(ctx, name)
	if err != nil {
		return nil, NewErrorf(InternalErr, err)
	}

	if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}
