package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// hookConfigUpdateFields whitelists fields allowed to be updated for HookConfigs
var hookConfigUpdateFields = []string{
	"Command",
	"Timeout",
	"Stdin",
}

// HookController exposes actions in which a viewer can perform.
type HookController struct {
	Store store.HookConfigStore
}

// NewHookController returns new HookController
func NewHookController(store store.HookConfigStore) HookController {
	return HookController{
		Store: store,
	}
}

// Query returns resources available to the viewer.
func (a HookController) Query(ctx context.Context) ([]*types.HookConfig, error) {
	// Fetch from store
	results, serr := a.Store.GetHookConfigs(ctx)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	return results, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (a HookController) Find(ctx context.Context, name string) (*types.HookConfig, error) {
	// Fetch from store
	result, serr := a.Store.GetHookConfigByName(ctx, name)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}
	if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}

// Create creates a Hook. If the Hook already exists, an error is returned.
func (a HookController) Create(ctx context.Context, newHook types.HookConfig) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &newHook)

	// Check for existing
	if e, err := a.Store.GetHookConfigByName(ctx, newHook.Name); err != nil {
		return NewError(InternalErr, err)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	// Validate
	if err := newHook.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.Store.UpdateHookConfig(ctx, &newHook); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// CreateOrReplace creates or replaces a Hook.
func (a HookController) CreateOrReplace(ctx context.Context, newHook types.HookConfig) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &newHook)

	// Validate
	if err := newHook.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.Store.UpdateHookConfig(ctx, &newHook); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Destroy removes a resource if viewer has access.
func (a HookController) Destroy(ctx context.Context, name string) error {
	// Fetch from store
	result, serr := a.Store.GetHookConfigByName(ctx, name)
	if serr != nil {
		return NewError(InternalErr, serr)
	} else if result == nil {
		return NewErrorf(NotFound)
	}

	// Remove from store
	if err := a.Store.DeleteHookConfigByName(ctx, result.Name); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}
