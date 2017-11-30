package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// checkConfigUpdateFields whitelists fields allowed to be updated for CheckConfigs
var entityUpdateFields = []string{
	"Subscriptions",
}

// EntityController exposes actions in which a viewer can perform.
type EntityController struct {
	Store  store.EntityStore
	Policy authorization.EntityPolicy
}

// NewEntityController returns new EntityController
func NewEntityController(store store.EntityStore) EntityController {
	return EntityController{
		Store:  store,
		Policy: authorization.Entities,
	}
}

// Destroy removes a resource if viewer has access.
func (c EntityController) Destroy(ctx context.Context, id string) error {
	abilities := c.Policy.WithContext(ctx)

	// Verify user has permission
	if yes := abilities.CanDelete(); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Fetch from store
	result, serr := c.Store.GetEntityByID(ctx, id)
	if serr != nil {
		return NewError(InternalErr, serr)
	} else if result == nil {
		return NewErrorf(NotFound)
	}

	// Remove from store
	if err := c.Store.DeleteEntityByID(ctx, result.ID); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (c EntityController) Find(ctx context.Context, id string) (*types.Entity, error) {
	// Fetch from store
	result, serr := c.Store.GetEntityByID(ctx, id)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	// Verify user has permission to view
	abilities := c.Policy.WithContext(ctx)
	if result != nil && abilities.CanRead(result) {
		return result, nil
	}

	return nil, NewErrorf(NotFound)
}

// Query returns resources available to the viewer.
func (c EntityController) Query(ctx context.Context) ([]*types.Entity, error) {
	// Fetch from store
	results, serr := c.Store.GetEntities(ctx)
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
func (c EntityController) Update(ctx context.Context, given types.Entity) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &given)
	abilities := c.Policy.WithContext(ctx)

	// Find existing entity
	entity, err := c.Store.GetEntityByID(ctx, given.ID)
	if err != nil {
		return NewError(InternalErr, err)
	} else if entity == nil {
		return NewErrorf(NotFound)
	}

	// Verify viewer can make change
	if yes := abilities.CanUpdate(entity); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Copy
	copyFields(entity, &given, entityUpdateFields...)

	// Validate
	if err := entity.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist Changes
	if serr := c.Store.UpdateEntity(ctx, entity); serr != nil {
		return NewError(InternalErr, serr)
	}

	return nil
}
