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

// GetClusterRoleBindingsPath gets the path of the cluster role binding store.
func GetClusterRoleBindingsPath(ctx context.Context, name string) string {
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
	return Delete(ctx, s.client, GetClusterRoleBindingsPath(ctx, name))
}

// GetClusterRoleBinding ...
func (s *Store) GetClusterRoleBinding(ctx context.Context, name string) (*types.ClusterRoleBinding, error) {
	role := &types.ClusterRoleBinding{}
	err := Get(ctx, s.client, GetClusterRoleBindingsPath(ctx, name), role)
	return role, err
}

// ListClusterRoleBindings ...
func (s *Store) ListClusterRoleBindings(ctx context.Context, pred *store.SelectionPredicate) ([]*types.ClusterRoleBinding, error) {
	if pred == nil {
		pred = &store.SelectionPredicate{}
	}
	roles := []*types.ClusterRoleBinding{}
	err := List(ctx, s.client, GetClusterRoleBindingsPath, &roles, pred)
	return roles, err
}

// UpdateClusterRoleBinding ...
func (s *Store) UpdateClusterRoleBinding(ctx context.Context, clusterRoleBinding *types.ClusterRoleBinding) error {
	if err := clusterRoleBinding.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return Update(ctx, s.client, getClusterRoleBindingPath(clusterRoleBinding), "", clusterRoleBinding)
}
