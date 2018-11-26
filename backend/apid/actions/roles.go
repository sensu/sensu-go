package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// RoleController exposes the Roles.
type RoleController struct {
	Store store.RoleStore
}

// NewRoleController creates a new RolesController.
func NewRoleController(store store.RoleStore) RoleController {
	return RoleController{
		Store: store,
	}
}

// Create creates a new role.
// Returns an error if the role already exists.
func (a RoleController) Create(ctx context.Context, role types.Role) error {
	if err := a.Store.CreateRole(ctx, &role); err != nil {
		switch err := err.(type) {
		case *store.ErrAlreadyExists:
			return NewErrorf(AlreadyExistsErr)
		case *store.ErrNotValid:
			return NewErrorf(InvalidArgument)
		default:
			return NewError(InternalErr, err)
		}
	}

	return nil
}

// CreateOrReplace creates or replaces a role.
func (a RoleController) CreateOrReplace(ctx context.Context, role types.Role) error {
	if err := a.Store.CreateOrUpdateRole(ctx, &role); err != nil {
		switch err := err.(type) {
		case *store.ErrNotValid:
			return NewErrorf(InvalidArgument)
		default:
			return NewError(InternalErr, err)
		}
	}

	return nil
}

// Destroy removes the given role from the store.
func (a RoleController) Destroy(ctx context.Context, name string) error {
	if err := a.Store.DeleteRole(ctx, name); err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return NewErrorf(NotFound)
		default:
			return NewError(InternalErr, err)
		}
	}

	return nil
}

// Get retrieves the role with the given name.
func (a RoleController) Get(ctx context.Context, name string) (*types.Role, error) {
	role, err := a.Store.GetRole(ctx, name)
	if err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return nil, NewErrorf(NotFound)
		default:
			return nil, NewError(InternalErr, err)
		}
	}

	return role, nil
}

// List returns all available roles.
func (a RoleController) List(ctx context.Context) ([]*types.Role, error) {
	// Fetch from store
	results, err := a.Store.ListRoles(ctx)
	if err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return nil, NewErrorf(NotFound)
		default:
			return nil, NewError(InternalErr, err)
		}
	}

	return results, nil
}
