package actions

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// RoleController exposes the Roles.
type RoleController struct {
	store store.RoleStore
}

// NewRoleController creates a new RolesController.
func NewRoleController(store store.RoleStore) RoleController {
	return RoleController{
		store: store,
	}
}

// Create creates a new role.
// Returns an error if the role already exists.
func (a RoleController) Create(ctx context.Context, role types.Role) error {
	if err := a.store.CreateRole(ctx, &role); err != nil {
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
	if err := a.store.CreateOrUpdateRole(ctx, &role); err != nil {
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
	if err := a.store.DeleteRole(ctx, name); err != nil {
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
	role, err := a.store.GetRole(ctx, name)
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
func (a RoleController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	// Fetch from store
	results, err := a.store.ListRoles(ctx, pred)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	resources := make([]corev2.Resource, len(results))
	for i, v := range results {
		resources[i] = corev2.Resource(v)
	}

	return resources, nil
}
