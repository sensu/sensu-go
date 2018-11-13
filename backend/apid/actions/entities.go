package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// entityUpdateFields whitelists fields allowed to be updated for Entities
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
	result, serr := c.Store.GetEntityByName(ctx, id)
	if serr != nil {
		return NewError(InternalErr, serr)
	} else if result == nil {
		return NewErrorf(NotFound)
	}

	// Remove from store
	if err := c.Store.DeleteEntityByName(ctx, result.Name); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (c EntityController) Find(ctx context.Context, id string) (*types.Entity, error) {
	// Fetch from store
	result, serr := c.Store.GetEntityByName(ctx, id)
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

// Create instatiates, validates and persists new resource if viewer has access.
func (c EntityController) Create(ctx context.Context, entity types.Entity) error {
	ctx = addOrgEnvToContext(ctx, &entity)
	abilities := c.Policy.WithContext(ctx)

	// Verify viewer can create the resource
	if yes := abilities.CanCreate(&entity); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Check for an already existing resource
	if e, err := c.Store.GetEntityByName(ctx, entity.Name); err != nil {
		return NewError(InternalErr, err)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	// Validate the resource
	if err := entity.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist the resource in the store
	if err := c.Store.UpdateEntity(ctx, &entity); err != nil {
		return NewError(InternalErr, err)
	}
	return nil
}

// CreateOrReplace creates or replaces an entity. It returns an error if the
// provided entity is invalid, the user doesn't have permissions to create or
// update the entity, or if an internal error is returned from the store.
func (c EntityController) CreateOrReplace(ctx context.Context, entity types.Entity) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &entity)
	abilities := c.Policy.WithContext(ctx)

	// Verify user permissions
	if !(abilities.CanCreate(&entity) && abilities.CanUpdate(&entity)) {
		return NewErrorf(PermissionDenied, "create/update")
	}

	// Validate
	if err := entity.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist Changes
	if serr := c.Store.UpdateEntity(ctx, &entity); serr != nil {
		return NewError(InternalErr, serr)
	}

	return nil
}

// Update validates and persists changes to a resource if viewer has access.
func (c EntityController) Update(ctx context.Context, given types.Entity) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &given)
	abilities := c.Policy.WithContext(ctx)

	// Find existing entity
	entity, err := c.Store.GetEntityByName(ctx, given.Name)
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
