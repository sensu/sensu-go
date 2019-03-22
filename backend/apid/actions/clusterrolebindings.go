package actions

import (
	"context"
	"encoding/base64"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// ClusterRoleBindingController exposes the ClusterRoleBindings.
type ClusterRoleBindingController struct {
	Store store.ClusterRoleBindingStore
}

// NewClusterRoleBindingController creates a new ClusterRoleBindingController.
func NewClusterRoleBindingController(store store.ClusterRoleBindingStore) ClusterRoleBindingController {
	return ClusterRoleBindingController{
		Store: store,
	}
}

// Create creates a new cluster role binding.
// Returns an error if the cluster role binding already exists.
func (a ClusterRoleBindingController) Create(ctx context.Context, role types.ClusterRoleBinding) error {
	if err := a.Store.CreateClusterRoleBinding(ctx, &role); err != nil {
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

// CreateOrReplace creates or replaces a cluster role binding.
func (a ClusterRoleBindingController) CreateOrReplace(ctx context.Context, role types.ClusterRoleBinding) error {
	if err := a.Store.CreateOrUpdateClusterRoleBinding(ctx, &role); err != nil {
		switch err := err.(type) {
		case *store.ErrNotValid:
			return NewErrorf(InvalidArgument)
		default:
			return NewError(InternalErr, err)
		}
	}

	return nil
}

// Destroy removes the given cluster role binding from the store.
func (a ClusterRoleBindingController) Destroy(ctx context.Context, name string) error {
	if err := a.Store.DeleteClusterRoleBinding(ctx, name); err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return NewErrorf(NotFound)
		default:
			return NewError(InternalErr, err)
		}
	}

	return nil
}

// Get retrieves the cluster role binding with the given name.
func (a ClusterRoleBindingController) Get(ctx context.Context, name string) (*types.ClusterRoleBinding, error) {
	role, err := a.Store.GetClusterRoleBinding(ctx, name)
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

// List returns all available cluster role bindings.
func (a ClusterRoleBindingController) List(ctx context.Context) ([]*types.ClusterRoleBinding, string, error) {
	pageSize := corev2.PageSizeFromContext(ctx)
	continueToken := corev2.PageContinueFromContext(ctx)

	// Fetch from store
	results, newContinueToken, err := a.Store.ListClusterRoleBindings(ctx, int64(pageSize), continueToken)
	if err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return nil, "", NewErrorf(NotFound)
		default:
			return nil, "", NewError(InternalErr, err)
		}
	}

	// Encode the continue token with base64url (RFC 4648), without padding
	encodedNewContinueToken := base64.RawURLEncoding.EncodeToString([]byte(newContinueToken))

	return results, encodedNewContinueToken, nil
}
