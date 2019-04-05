package actions

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// NamespacesController defines the fields required for this controller.
type NamespacesController struct {
	store store.NamespaceStore
}

// NewNamespacesController returns new NamespacesController
func NewNamespacesController(store store.NamespaceStore) NamespacesController {
	return NamespacesController{
		store: store,
	}
}

// List returns namespaces
func (a NamespacesController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	// Fetch from store
	results, err := a.store.ListNamespaces(ctx, pred)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	resources := make([]corev2.Resource, len(results))
	for i, v := range results {
		resources[i] = corev2.Resource(v)
	}

	return resources, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (a NamespacesController) Find(ctx context.Context, name string) (*types.Namespace, error) {
	// Fetch from store
	result, serr := a.store.GetNamespace(ctx, name)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}
	if result == nil {
		return nil, NewErrorf(NotFound)
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
	if err := a.store.CreateNamespace(ctx, &namespace); err != nil {
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
	if err := a.store.UpdateNamespace(ctx, &namespace); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Destroy removes a resource if viewer has access.
func (a NamespacesController) Destroy(ctx context.Context, name string) error {
	// Fetch from store
	result, serr := a.store.GetNamespace(ctx, name)
	if serr != nil {
		return NewError(InternalErr, serr)
	} else if result == nil {
		return NewErrorf(NotFound)
	}

	// Remove from store
	if err := a.store.DeleteNamespace(ctx, result.Name); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}
