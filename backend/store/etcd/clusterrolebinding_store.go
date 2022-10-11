package etcd

import (
	"context"
	"path"

	"github.com/sensu/sensu-go/backend/store"
	corev2 "github.com/sensu/core/v2"
)

var (
	clusterRoleBindingPathPrefix = "rbac/clusterrolebindings"
)

func getClusterRoleBindingPath(clusterRole *corev2.ClusterRoleBinding) string {
	return path.Join(store.Root, clusterRoleBindingPathPrefix, clusterRole.Name)
}

// GetClusterRoleBindingsPath gets the path of the cluster role binding store.
func GetClusterRoleBindingsPath(ctx context.Context, name string) string {
	return path.Join(store.Root, clusterRoleBindingPathPrefix, name)
}

// CreateClusterRoleBinding ...
func (s *Store) CreateClusterRoleBinding(ctx context.Context, clusterRoleBinding *corev2.ClusterRoleBinding) error {
	if err := clusterRoleBinding.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return Create(ctx, s.client, getClusterRoleBindingPath(clusterRoleBinding), "", clusterRoleBinding)
}

// CreateOrUpdateClusterRoleBinding ...
func (s *Store) CreateOrUpdateClusterRoleBinding(ctx context.Context, clusterRoleBinding *corev2.ClusterRoleBinding) error {
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
func (s *Store) GetClusterRoleBinding(ctx context.Context, name string) (*corev2.ClusterRoleBinding, error) {
	role := &corev2.ClusterRoleBinding{}
	err := Get(ctx, s.client, GetClusterRoleBindingsPath(ctx, name), role)
	return role, err
}

// ListClusterRoleBindings ...
func (s *Store) ListClusterRoleBindings(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.ClusterRoleBinding, error) {
	roles := []*corev2.ClusterRoleBinding{}
	err := List(ctx, s.client, GetClusterRoleBindingsPath, &roles, pred)
	return roles, err
}

// UpdateClusterRoleBinding ...
func (s *Store) UpdateClusterRoleBinding(ctx context.Context, clusterRoleBinding *corev2.ClusterRoleBinding) error {
	if err := clusterRoleBinding.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return Update(ctx, s.client, getClusterRoleBindingPath(clusterRoleBinding), "", clusterRoleBinding)
}
