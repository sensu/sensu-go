package etcd

import (
	"context"
	"path"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var (
	clusterRolesPathPrefix = "rbac/clusterroles"
)

func getClusterRolePath(clusterRole *types.ClusterRole) string {
	return path.Join(store.Root, clusterRolesPathPrefix, clusterRole.Name)
}

func getClusterRolesPath(ctx context.Context, name string) string {
	return path.Join(store.Root, clusterRolesPathPrefix, name)
}

// CreateClusterRole ...
func (s *Store) CreateClusterRole(ctx context.Context, clusterRole *types.ClusterRole) error {
	if err := clusterRole.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return Create(ctx, s.client, getClusterRolePath(clusterRole), "", clusterRole)
}

// CreateOrUpdateClusterRole ...
func (s *Store) CreateOrUpdateClusterRole(ctx context.Context, clusterRole *types.ClusterRole) error {
	if err := clusterRole.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return CreateOrUpdate(ctx, s.client, getClusterRolePath(clusterRole), "", clusterRole)
}

// DeleteClusterRole ...
func (s *Store) DeleteClusterRole(ctx context.Context, name string) error {
	return Delete(ctx, s.client, getClusterRolesPath(ctx, name))
}

// GetClusterRole ...
func (s *Store) GetClusterRole(ctx context.Context, name string) (*types.ClusterRole, error) {
	clusterRole := &types.ClusterRole{}
	err := Get(ctx, s.client, getClusterRolesPath(ctx, name), clusterRole)
	return clusterRole, err
}

// ListClusterRoles ...
func (s *Store) ListClusterRoles(ctx context.Context, pageSize int64, continueToken string) ([]*types.ClusterRole, string, error) {
	clusterRoles := []*types.ClusterRole{}
	nextContinueToken, err := List(ctx, s.client, getClusterRolesPath, &clusterRoles, pageSize, continueToken)
	return clusterRoles, nextContinueToken, err
}

// UpdateClusterRole ...
func (s *Store) UpdateClusterRole(ctx context.Context, clusterRole *types.ClusterRole) error {
	if err := clusterRole.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return Update(ctx, s.client, getClusterRolePath(clusterRole), "", clusterRole)
}
