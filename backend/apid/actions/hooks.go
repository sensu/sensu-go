package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/authorization"
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
	Store  store.HookConfigStore
	Policy authorization.HookPolicy
}

// NewHookController returns new HookController
func NewHookController(store store.HookConfigStore) HookController {
	return HookController{
		Store:  store,
		Policy: authorization.Hooks,
	}
}

// Query returns resources available to the viewer.
func (a HookController) Query(ctx context.Context) ([]*types.HookConfig, error) {
	// Fetch from store
	results, serr := a.Store.GetHookConfigs(ctx)
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
func (a HookController) Find(ctx context.Context, name string) (*types.HookConfig, error) {
	// Fetch from store
	result, serr := a.Store.GetHookConfigByName(ctx, name)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	// Verify user has permission to view
	abilities := a.Policy.WithContext(ctx)
	if result != nil && abilities.CanRead(result) {
		return result, nil
	}

	return nil, NewErrorf(NotFound)
}

// Create creates a Hook. If the Hook already exists, an error is returned.
func (a HookController) Create(ctx context.Context, newHook types.HookConfig) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &newHook)
	abilities := a.Policy.WithContext(ctx)

	// Check for existing
	if e, err := a.Store.GetHookConfigByName(ctx, newHook.Name); err != nil {
		return NewError(InternalErr, err)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	// Verify viewer can make change
	if yes := abilities.CanCreate(&newHook); !yes {
		return NewErrorf(PermissionDenied)
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
	abilities := a.Policy.WithContext(ctx)

	// Verify viewer can make change
	if !(abilities.CanCreate(&newHook) && abilities.CanUpdate(&newHook)) {
		return NewErrorf(PermissionDenied)
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

// Update validates and persists changes to a resource if viewer has access.
func (a HookController) Update(ctx context.Context, given types.HookConfig) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &given)
	abilities := a.Policy.WithContext(ctx)

	// Find existing hook
	hook, err := a.Store.GetHookConfigByName(ctx, given.Name)
	if err != nil {
		return NewError(InternalErr, err)
	} else if hook == nil {
		return NewErrorf(NotFound)
	}

	// Verify viewer can make change
	if yes := abilities.CanUpdate(hook); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Copy
	copyFields(hook, &given, hookConfigUpdateFields...)

	// Validate
	if err := hook.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist Changes
	if serr := a.Store.UpdateHookConfig(ctx, hook); serr != nil {
		return NewError(InternalErr, serr)
	}

	return nil
}

// Destroy removes a resource if viewer has access.
func (a HookController) Destroy(ctx context.Context, name string) error {
	abilities := a.Policy.WithContext(ctx)

	// Verify user has permission
	if yes := abilities.CanDelete(); !yes {
		return NewErrorf(PermissionDenied)
	}

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
