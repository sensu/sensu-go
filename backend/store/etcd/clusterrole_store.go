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
	return s.create(ctx, getClusterRolePath(clusterRole), "", clusterRole)
}

// CreateOrUpdateClusterRole ...
func (s *Store) CreateOrUpdateClusterRole(ctx context.Context, clusterRole *types.ClusterRole) error {
	if err := clusterRole.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return s.createOrUpdate(ctx, getClusterRolePath(clusterRole), "", clusterRole)
}

// DeleteClusterRole ...
func (s *Store) DeleteClusterRole(ctx context.Context, name string) error {
	return s.delete(ctx, getClusterRolesPath(ctx, name))
}

// GetClusterRole ...
func (s *Store) GetClusterRole(ctx context.Context, name string) (*types.ClusterRole, error) {
	clusterRole := &types.ClusterRole{}
	err := s.get(ctx, getClusterRolesPath(ctx, name), clusterRole)
	return clusterRole, err
}

// ListClusterRoles ...
func (s *Store) ListClusterRoles(ctx context.Context) ([]*types.ClusterRole, error) {
	clusterRoles := []*types.ClusterRole{}
	err := s.list(ctx, getClusterRolesPath, &clusterRoles)
	return clusterRoles, err
}

// UpdateClusterRole ...
func (s *Store) UpdateClusterRole(ctx context.Context, clusterRole *types.ClusterRole) error {
	if err := clusterRole.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return s.update(ctx, getClusterRolePath(clusterRole), "", clusterRole)
}
