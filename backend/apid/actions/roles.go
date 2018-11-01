package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// roleUpdateFields refers to fields a viewer may update
var roleUpdateFields = []string{"Rules"}

// RoleController exposes the Roles
type RoleController struct {
	Store store.RoleStore
}

// NewRoleController initializes a RoleController
func NewRoleController(store store.RoleStore) RoleController {
	return RoleController{
		Store: store,
	}
}

// Query returns resources available to the viewer filter by given params.
func (a RoleController) Query(ctx context.Context) ([]*types.Role, error) {
	// Fetch from store
	results, serr := a.Store.ListRoles(ctx)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	return results, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (a RoleController) Find(ctx context.Context, name string) (*types.Role, error) {
	// Fetch from store
	result, serr := a.findRole(ctx, name)
	if serr != nil {
		return nil, serr
	}

	return result, nil
}

// Create creates a new role. It returns an error if the role already exists.
func (a RoleController) Create(ctx context.Context, newRole types.Role) error {
	// Role for existing
	if e, err := a.Store.GetRole(ctx, newRole.Name); err != nil {
		return NewError(InternalErr, err)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	// Validate
	if err := newRole.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.Store.UpdateRole(ctx, &newRole); err != nil {
		return NewError(InternalErr, err)
	}
	return nil
}

// CreateOrReplace creates or replaces a role.
func (a RoleController) CreateOrReplace(ctx context.Context, newRole types.Role) error {
	// Validate
	if err := newRole.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.Store.UpdateRole(ctx, &newRole); err != nil {
		return NewError(InternalErr, err)
	}
	return nil
}

// Update validates and persists changes to a resource if viewer has access.
func (a RoleController) Update(ctx context.Context, given types.Role) error {
	return a.findAndUpdateRole(ctx, given.Name, func(role *types.Role) error {
		copyFields(role, &given, roleUpdateFields...)
		return nil
	})
}

// Destroy removes given role from the store.
func (a RoleController) Destroy(ctx context.Context, name string) error {
	// Fetch from store
	_, err := a.findRole(ctx, name)
	if err != nil {
		return err
	}
	// Remove from store
	if serr := a.Store.DeleteRole(ctx, name); serr != nil {
		return NewError(InternalErr, serr)
	}
	return nil
}

func (a RoleController) findAndUpdateRole(
	ctx context.Context,
	name string,
	configureFn func(*types.Role) error,
) error {
	// Find
	role, serr := a.findRole(ctx, name)
	if serr != nil {
		return serr
	}

	// Configure
	if err := configureFn(role); err != nil {
		return err
	}
	// Update
	return a.updateRole(ctx, role)
}

func (a RoleController) findRole(ctx context.Context, name string) (*types.Role, error) {
	result, serr := a.Store.GetRole(ctx, name)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	} else if result == nil {
		return nil, NewErrorf(NotFound)
	}
	return result, nil
}

func (a RoleController) updateRole(ctx context.Context, role *types.Role) error {
	if err := role.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}
	if err := a.Store.UpdateRole(ctx, role); err != nil {
		return NewError(InternalErr, err)
	}
	return nil
}
