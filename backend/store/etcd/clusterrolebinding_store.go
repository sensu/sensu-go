package etcd

import (
	"context"
	"path"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var (
	clusterRoleBindingPathPrefix = "rbac/clusterrolebindings"
)

func getClusterRoleBindingPath(clusterRole *types.ClusterRoleBinding) string {
	return path.Join(store.Root, clusterRoleBindingPathPrefix, clusterRole.Name)
}

func getClusterRoleBindingsPath(ctx context.Context, name string) string {
	return path.Join(store.Root, clusterRoleBindingPathPrefix, name)
}

// CreateClusterRoleBinding ...
func (s *Store) CreateClusterRoleBinding(ctx context.Context, clusterRoleBinding *types.ClusterRoleBinding) error {
	if err := clusterRoleBinding.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return Create(ctx, s.client, getClusterRoleBindingPath(clusterRoleBinding), "", clusterRoleBinding)
}

// CreateOrUpdateClusterRoleBinding ...
func (s *Store) CreateOrUpdateClusterRoleBinding(ctx context.Context, clusterRoleBinding *types.ClusterRoleBinding) error {
	if err := clusterRoleBinding.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return CreateOrUpdate(ctx, s.client, getClusterRoleBindingPath(clusterRoleBinding), "", clusterRoleBinding)
}

// DeleteClusterRoleBinding ...
func (s *Store) DeleteClusterRoleBinding(ctx context.Context, name string) error {
	return Delete(ctx, s.client, getClusterRoleBindingsPath(ctx, name))
}

// GetClusterRoleBinding ...
func (s *Store) GetClusterRoleBinding(ctx context.Context, name string) (*types.ClusterRoleBinding, error) {
	role := &types.ClusterRoleBinding{}
	err := Get(ctx, s.client, getClusterRoleBindingsPath(ctx, name), role)
	return role, err
}

// ListClusterRoleBindings ...
func (s *Store) ListClusterRoleBindings(ctx context.Context) ([]*types.ClusterRoleBinding, error) {
	roles := []*types.ClusterRoleBinding{}
	err := List(ctx, s.client, getClusterRoleBindingsPath, &roles)
	return roles, err
}

// UpdateClusterRoleBinding ...
func (s *Store) UpdateClusterRoleBinding(ctx context.Context, clusterRoleBinding *types.ClusterRoleBinding) error {
	if err := clusterRoleBinding.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return Update(ctx, s.client, getClusterRoleBindingPath(clusterRoleBinding), "", clusterRoleBinding)
}
