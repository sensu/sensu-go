package etcd

import (
	"context"
	"path"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

var (
	clusterRolesPathPrefix = "rbac/clusterroles"
)

func getClusterRolePath(clusterRole *v2.ClusterRole) string {
	return path.Join(store.Root, clusterRolesPathPrefix, clusterRole.Name)
}

// GetClusterRolesPath gets the path of the cluster role store.
func GetClusterRolesPath(ctx context.Context, name string) string {
	return path.Join(store.Root, clusterRolesPathPrefix, name)
}

// CreateClusterRole ...
func (s *Store) CreateClusterRole(ctx context.Context, clusterRole *v2.ClusterRole) error {
	if err := clusterRole.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return Create(ctx, s.client, getClusterRolePath(clusterRole), "", clusterRole)
}

// CreateOrUpdateClusterRole ...
func (s *Store) CreateOrUpdateClusterRole(ctx context.Context, clusterRole *v2.ClusterRole) error {
	if err := clusterRole.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return CreateOrUpdate(ctx, s.client, getClusterRolePath(clusterRole), "", clusterRole)
}

// DeleteClusterRole ...
func (s *Store) DeleteClusterRole(ctx context.Context, name string) error {
	return Delete(ctx, s.client, GetClusterRolesPath(ctx, name))
}

// GetClusterRole ...
func (s *Store) GetClusterRole(ctx context.Context, name string) (*v2.ClusterRole, error) {
	clusterRole := &v2.ClusterRole{}
	err := Get(ctx, s.client, GetClusterRolesPath(ctx, name), clusterRole)
	return clusterRole, err
}

// ListClusterRoles ...
func (s *Store) ListClusterRoles(ctx context.Context, pred *store.SelectionPredicate) ([]*v2.ClusterRole, error) {
	clusterRoles := []*v2.ClusterRole{}
	err := List(ctx, s.client, GetClusterRolesPath, &clusterRoles, pred)
	return clusterRoles, err
}

// UpdateClusterRole ...
func (s *Store) UpdateClusterRole(ctx context.Context, clusterRole *v2.ClusterRole) error {
	if err := clusterRole.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return Update(ctx, s.client, getClusterRolePath(clusterRole), "", clusterRole)
}
