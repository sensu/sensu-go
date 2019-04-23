package actions

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
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
	store store.EventFilterStore
}

// NewEventFilterController creates a new EventFilterController backed by store.
func NewEventFilterController(store store.EventFilterStore) EventFilterController {
	return EventFilterController{
		store: store,
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
	if m, err := c.store.GetEventFilterByName(ctx, filter.Name); err != nil {
		return NewError(InternalErr, err)
	} else if m != nil {
		return NewErrorf(AlreadyExistsErr, filter.Name)
	}

	// Validate
	if err := filter.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := c.store.UpdateEventFilter(ctx, &filter); err != nil {
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
	if err := c.store.UpdateEventFilter(ctx, &filter); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// List returns event filters
func (c EventFilterController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	// Fetch from store
	results, err := c.store.GetEventFilters(ctx, pred)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	resources := make([]corev2.Resource, len(results))
	for i, v := range results {
		resources[i] = corev2.Resource(v)
	}

	return resources, nil
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
	filter, err := c.store.GetEventFilterByName(ctx, name)
	if err != nil {
		return NewError(InternalErr, err)
	}
	if filter == nil {
		return NewErrorf(NotFound, name)
	}

	// Remove from store
	if err := c.store.DeleteEventFilterByName(ctx, filter.Name); err != nil {
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

	result, err := c.store.GetEventFilterByName(ctx, name)
	if err != nil {
		return nil, NewErrorf(InternalErr, err)
	}

	if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}
