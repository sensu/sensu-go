package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// NamespacesController defines the fields required for this controller.
type NamespacesController struct {
	Store  store.NamespaceStore
	Policy authorization.NamespacePolicy
}

// NewNamespacesController returns new NamespacesController
func NewNamespacesController(store store.NamespaceStore) NamespacesController {
	return NamespacesController{
		Store:  store,
		Policy: authorization.Namespaces,
	}
}

// Query returns resources available to the viewer filter by given params.
func (a NamespacesController) Query(ctx context.Context) ([]*types.Namespace, error) {
	// Fetch from store
	results, serr := a.Store.ListNamespaces(ctx)
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
func (a NamespacesController) Find(ctx context.Context, name string) (*types.Namespace, error) {
	// Fetch from store
	result, serr := a.Store.GetNamespace(ctx, name)
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

// Create creates a new namespace. It returns an error if the  namespace exists.
func (a NamespacesController) Create(ctx context.Context, namespace types.Namespace) error {
	abilities := a.Policy.WithContext(ctx)

	// Verify viewer can make change
	if yes := abilities.CanCreate(&namespace); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Validate
	if err := namespace.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.Store.CreateNamespace(ctx, &namespace); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// CreateOrReplace creates or replaces an namespace.
func (a NamespacesController) CreateOrReplace(ctx context.Context, namespace types.Namespace) error {
	abilities := a.Policy.WithContext(ctx)

	// Verify viewer can make change
	if !(abilities.CanCreate(&namespace) && abilities.CanUpdate(&namespace)) {
		return NewErrorf(PermissionDenied)
	}

	// Validate
	if err := namespace.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.Store.UpdateNamespace(ctx, &namespace); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Update validates and persists changes to a resource if viewer has access.
func (a NamespacesController) Update(ctx context.Context, given types.Namespace) error {
	abilities := a.Policy.WithContext(ctx)

	// Find existing namespace
	namespace, err := a.Store.GetNamespace(ctx, given.Name)
	if err != nil {
		return NewError(InternalErr, err)
	} else if namespace == nil {
		return NewErrorf(NotFound)
	}

	// Verify viewer can make change
	if yes := abilities.CanUpdate(namespace); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Validate
	if err := namespace.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist Changes
	if serr := a.Store.UpdateNamespace(ctx, namespace); serr != nil {
		return NewError(InternalErr, serr)
	}

	return nil
}

// Destroy removes a resource if viewer has access.
func (a NamespacesController) Destroy(ctx context.Context, name string) error {
	abilities := a.Policy.WithContext(ctx)

	// Verify user has permission
	if yes := abilities.CanDelete(); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Fetch from store
	result, serr := a.Store.GetNamespace(ctx, name)
	if serr != nil {
		return NewError(InternalErr, serr)
	} else if result == nil {
		return NewErrorf(NotFound)
	}

	// Remove from store
	if err := a.Store.DeleteNamespace(ctx, result.Name); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}
