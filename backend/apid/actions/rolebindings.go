package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// RoleBindingController exposes the Roles.
type RoleBindingController struct {
	Store store.RoleBindingStore
}

// NewRoleBindingController creates a new RoleBindingController.
func NewRoleBindingController(store store.RoleBindingStore) RoleBindingController {
	return RoleBindingController{
		Store: store,
	}
}

// Create creates a new role binding.
// Returns an error if the role binding already exists.
func (a RoleBindingController) Create(ctx context.Context, role types.RoleBinding) error {
	if err := a.Store.CreateRoleBinding(ctx, &role); err != nil {
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

// CreateOrReplace creates or replaces a role binding.
func (a RoleBindingController) CreateOrReplace(ctx context.Context, role types.RoleBinding) error {
	if err := a.Store.CreateOrUpdateRoleBinding(ctx, &role); err != nil {
		switch err := err.(type) {
		case *store.ErrNotValid:
			return NewErrorf(InvalidArgument)
		default:
			return NewError(InternalErr, err)
		}
	}

	return nil
}

// Destroy removes the given role binding from the store.
func (a RoleBindingController) Destroy(ctx context.Context, name string) error {
	if err := a.Store.DeleteRoleBinding(ctx, name); err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return NewErrorf(NotFound)
		default:
			return NewError(InternalErr, err)
		}
	}

	return nil
}

// Get retrieves the role binding with the given name.
func (a RoleBindingController) Get(ctx context.Context, name string) (*types.RoleBinding, error) {
	role, err := a.Store.GetRoleBinding(ctx, name)
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

// List returns all available role bindings.
func (a RoleBindingController) List(ctx context.Context) ([]*types.RoleBinding, error) {
	// Fetch from store
	results, err := a.Store.ListRoleBindings(ctx)
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
