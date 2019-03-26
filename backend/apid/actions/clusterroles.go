package actions

import (
	"context"
	"encoding/base64"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// ClusterRoleController exposes the ClusterRoles.
type ClusterRoleController struct {
	Store store.ClusterRoleStore
}

// NewClusterRoleController creates a new ClusterRoleController.
func NewClusterRoleController(store store.ClusterRoleStore) ClusterRoleController {
	return ClusterRoleController{
		Store: store,
	}
}

// Create creates a new cluster role.
// Returns an error if the cluster role already exists.
func (a ClusterRoleController) Create(ctx context.Context, role types.ClusterRole) error {
	if err := a.Store.CreateClusterRole(ctx, &role); err != nil {
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

// CreateOrReplace creates or replaces a cluster role.
func (a ClusterRoleController) CreateOrReplace(ctx context.Context, role types.ClusterRole) error {
	if err := a.Store.CreateOrUpdateClusterRole(ctx, &role); err != nil {
		switch err := err.(type) {
		case *store.ErrNotValid:
			return NewErrorf(InvalidArgument)
		default:
			return NewError(InternalErr, err)
		}
	}

	return nil
}

// Destroy removes the given cluster role from the store.
func (a ClusterRoleController) Destroy(ctx context.Context, name string) error {
	if err := a.Store.DeleteClusterRole(ctx, name); err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return NewErrorf(NotFound)
		default:
			return NewError(InternalErr, err)
		}
	}

	return nil
}

// Get retrieves the cluster role with the given name.
func (a ClusterRoleController) Get(ctx context.Context, name string) (*types.ClusterRole, error) {
	role, err := a.Store.GetClusterRole(ctx, name)
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

// List returns all available cluster roles.
func (a ClusterRoleController) List(ctx context.Context) ([]*types.ClusterRole, string, error) {
	pageSize := corev2.PageSizeFromContext(ctx)
	continueToken := corev2.PageContinueFromContext(ctx)

	// Fetch from store
	results, newContinueToken, err := a.Store.ListClusterRoles(ctx, int64(pageSize), continueToken)
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
