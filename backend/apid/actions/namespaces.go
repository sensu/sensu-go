package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// NamespacesController defines the fields required for this controller.
type NamespacesController struct {
	Store store.NamespaceStore
}

// NewNamespacesController returns new NamespacesController
func NewNamespacesController(store store.NamespaceStore) NamespacesController {
	return NamespacesController{
		Store: store,
	}
}

// Query returns resources available to the viewer filter by given params.
func (a NamespacesController) Query(ctx context.Context) ([]*types.Namespace, error) {
	// Fetch from store
	results, serr := a.Store.ListNamespaces(ctx)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
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

	return result, nil
}

// Create creates a new namespace. It returns an error if the  namespace exists.
func (a NamespacesController) Create(ctx context.Context, namespace types.Namespace) error {
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
	// Find existing namespace
	namespace, err := a.Store.GetNamespace(ctx, given.Name)
	if err != nil {
		return NewError(InternalErr, err)
	} else if namespace == nil {
		return NewErrorf(NotFound)
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
