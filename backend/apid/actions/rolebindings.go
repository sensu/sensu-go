package actions

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// RoleBindingController exposes the Roles.
type RoleBindingController struct {
	store store.RoleBindingStore
}

// NewRoleBindingController creates a new RoleBindingController.
func NewRoleBindingController(store store.RoleBindingStore) RoleBindingController {
	return RoleBindingController{
		store: store,
	}
}

// Create creates a new role binding.
// Returns an error if the role binding already exists.
func (a RoleBindingController) Create(ctx context.Context, role types.RoleBinding) error {
	if err := a.store.CreateRoleBinding(ctx, &role); err != nil {
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
	if err := a.store.CreateOrUpdateRoleBinding(ctx, &role); err != nil {
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
	if err := a.store.DeleteRoleBinding(ctx, name); err != nil {
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
	role, err := a.store.GetRoleBinding(ctx, name)
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
func (a RoleBindingController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	// Fetch from store
	results, err := a.store.ListRoleBindings(ctx, pred)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	resources := make([]corev2.Resource, len(results))
	for i, v := range results {
		resources[i] = corev2.Resource(v)
	}

	return resources, nil
}
