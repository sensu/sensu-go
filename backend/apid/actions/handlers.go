package actions

import (
	"context"
	"errors"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var updateFields = []string{
	"Filters",
	"Mutator",
	"Timeout",
	"Type",
	"Command",
	"Handlers",
	"Socket",
	"RuntimeAssets",
}

// HandlerController exposes actions available for handlers
type HandlerController struct {
	store store.HandlerStore
}

// NewHandlerController creates a new HandlerController backed by store.
func NewHandlerController(store store.HandlerStore) HandlerController {
	return HandlerController{
		store: store,
	}
}

// Create creates a new handler resource.
// It returns non-nil error if the new handler is invalid, update permissions
// do not exist, or an internal error occurs while updating the underlying
// Store.
func (c HandlerController) Create(ctx context.Context, handler types.Handler) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &handler)

	// Check for existing
	if m, err := c.store.GetHandlerByName(ctx, handler.Name); err != nil {
		return NewError(InternalErr, err)
	} else if m != nil {
		return NewErrorf(AlreadyExistsErr, handler.Name)
	}

	// Validate
	if err := handler.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	if handler.Type == types.HandlerGRPCType {
		return NewError(InvalidArgument, errors.New("use the extensions API for this handler type"))
	}

	// Persist
	if err := c.store.UpdateHandler(ctx, &handler); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// CreateOrReplace creates or replaces a handler resource.
// It returns non-nil error if the handler is invalid, permissions
// do not exist, or an internal error occurs while updating the underlying
// Store.
func (c HandlerController) CreateOrReplace(ctx context.Context, handler types.Handler) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &handler)

	// Validate
	if err := handler.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	if handler.Type == types.HandlerGRPCType {
		return NewError(InvalidArgument, errors.New("use the extensions API for this handler type"))
	}

	// Persist
	if err := c.store.UpdateHandler(ctx, &handler); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Destroy removes a resource if viewer has access.
func (c HandlerController) Destroy(ctx context.Context, name string) error {
	// Fetch from store
	result, serr := c.store.GetHandlerByName(ctx, name)
	if serr != nil {
		return NewError(InternalErr, serr)
	} else if result == nil {
		return NewErrorf(NotFound)
	}

	// Remove from store
	if err := c.store.DeleteHandlerByName(ctx, result.Name); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (c HandlerController) Find(ctx context.Context, name string) (*types.Handler, error) {
	// Fetch from store
	result, err := c.store.GetHandlerByName(ctx, name)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}
	if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}

// List returns resources available to the viewer
func (c HandlerController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	// Fetch from store
	results, err := c.store.GetHandlers(ctx, pred)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}
	resources := make([]corev2.Resource, len(results))
	for i, v := range results {
		resources[i] = corev2.Resource(v)
	}

	return resources, nil
}
