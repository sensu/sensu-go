package actions

import (
	"context"
	"errors"

	"github.com/sensu/sensu-go/backend/authorization"
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
}

// HandlerController exposes actions available for handlers
type HandlerController struct {
	Store  store.HandlerStore
	Policy authorization.HandlerPolicy
}

// NewHandlerController creates a new HandlerController backed by store.
func NewHandlerController(store store.HandlerStore) HandlerController {
	return HandlerController{
		Store:  store,
		Policy: authorization.Handlers,
	}
}

// Create creates a new handler resource.
// It returns non-nil error if the new handler is invalid, update permissions
// do not exist, or an internal error occurs while updating the underlying
// Store.
func (c HandlerController) Create(ctx context.Context, handler types.Handler) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &handler)
	abilities := c.Policy.WithContext(ctx)

	// Check for existing
	if m, err := c.Store.GetHandlerByName(ctx, handler.Name); err != nil {
		return NewError(InternalErr, err)
	} else if m != nil {
		return NewErrorf(AlreadyExistsErr, handler.Name)
	}

	// Verify permissions
	if ok := abilities.CanCreate(&handler); !ok {
		return NewErrorf(PermissionDenied, "create")
	}

	// Validate
	if err := handler.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	if handler.Type == types.HandlerGRPCType {
		return NewError(InvalidArgument, errors.New("use the extensions API for this handler type"))
	}

	// Persist
	if err := c.Store.UpdateHandler(ctx, &handler); err != nil {
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
	abilities := c.Policy.WithContext(ctx)

	// Verify permissions
	if !(abilities.CanCreate(&handler) && abilities.CanUpdate(&handler)) {
		return NewErrorf(PermissionDenied, "create/update")
	}

	// Validate
	if err := handler.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	if handler.Type == types.HandlerGRPCType {
		return NewError(InvalidArgument, errors.New("use the extensions API for this handler type"))
	}

	// Persist
	if err := c.Store.UpdateHandler(ctx, &handler); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Destroy removes a resource if viewer has access.
func (c HandlerController) Destroy(ctx context.Context, name string) error {
	abilities := c.Policy.WithContext(ctx)

	// Verify user has permission
	if yes := abilities.CanDelete(); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Fetch from store
	result, serr := c.Store.GetHandlerByName(ctx, name)
	if serr != nil {
		return NewError(InternalErr, serr)
	} else if result == nil {
		return NewErrorf(NotFound)
	}

	// Remove from store
	if err := c.Store.DeleteHandlerByName(ctx, result.Name); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (c HandlerController) Find(ctx context.Context, name string) (*types.Handler, error) {
	// Fetch from store
	result, err := c.Store.GetHandlerByName(ctx, name)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	// Verify user has permission to view
	abilities := c.Policy.WithContext(ctx)
	if result != nil && abilities.CanRead(result) {
		return result, nil
	}

	return nil, NewErrorf(NotFound)
}

// Query returns resources available to the viewer
func (c HandlerController) Query(ctx context.Context) ([]*types.Handler, error) {
	// Fetch from store
	results, serr := c.Store.GetHandlers(ctx)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	// Filter out those resources the viewer does not have access to view.
	abilities := c.Policy.WithContext(ctx)
	for i := 0; i < len(results); i++ {
		if !abilities.CanRead(results[i]) {
			results = append(results[:i], results[i+1:]...)
			i--
		}
	}

	return results, nil
}

// Update validates and persists changes to a resource if viewer has access.
func (c HandlerController) Update(ctx context.Context, newHandler types.Handler) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &newHandler)
	abilities := c.Policy.WithContext(ctx)

	// Find existing check
	handler, err := c.Store.GetHandlerByName(ctx, newHandler.Name)
	if err != nil {
		return NewError(InternalErr, err)
	} else if handler == nil {
		return NewErrorf(NotFound)
	}

	// Verify viewer can make change
	if yes := abilities.CanUpdate(handler); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Copy
	copyFields(handler, &newHandler, updateFields...)

	// Validate
	if err := handler.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist Changes
	if serr := c.Store.UpdateHandler(ctx, handler); serr != nil {
		return NewError(InternalErr, serr)
	}

	return nil
}
