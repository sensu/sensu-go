package etcd

import (
	"context"
	"path"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

var (
	clusterRoleBindingPathPrefix = "rbac/clusterrolebindings"
)

func getClusterRoleBindingPath(clusterRole *v2.ClusterRoleBinding) string {
	return path.Join(store.Root, clusterRoleBindingPathPrefix, clusterRole.Name)
}

// GetClusterRoleBindingsPath gets the path of the cluster role binding store.
func GetClusterRoleBindingsPath(ctx context.Context, name string) string {
	return path.Join(store.Root, clusterRoleBindingPathPrefix, name)
}

// CreateClusterRoleBinding ...
func (s *Store) CreateClusterRoleBinding(ctx context.Context, clusterRoleBinding *v2.ClusterRoleBinding) error {
	if err := clusterRoleBinding.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return Create(ctx, s.client, getClusterRoleBindingPath(clusterRoleBinding), "", clusterRoleBinding)
}

// CreateOrUpdateClusterRoleBinding ...
func (s *Store) CreateOrUpdateClusterRoleBinding(ctx context.Context, clusterRoleBinding *v2.ClusterRoleBinding) error {
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
func (s *Store) GetClusterRoleBinding(ctx context.Context, name string) (*v2.ClusterRoleBinding, error) {
	role := &v2.ClusterRoleBinding{}
	err := Get(ctx, s.client, GetClusterRoleBindingsPath(ctx, name), role)
	return role, err
}

// ListClusterRoleBindings ...
func (s *Store) ListClusterRoleBindings(ctx context.Context, pred *store.SelectionPredicate) ([]*v2.ClusterRoleBinding, error) {
	roles := []*v2.ClusterRoleBinding{}
	err := List(ctx, s.client, GetClusterRoleBindingsPath, &roles, pred)
	return roles, err
}

// UpdateClusterRoleBinding ...
func (s *Store) UpdateClusterRoleBinding(ctx context.Context, clusterRoleBinding *v2.ClusterRoleBinding) error {
	if err := clusterRoleBinding.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return Update(ctx, s.client, getClusterRoleBindingPath(clusterRoleBinding), "", clusterRoleBinding)
}
