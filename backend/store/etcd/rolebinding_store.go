package etcd

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var (
	roleBindingsPathPrefix = "rbac/rolebindings"
	roleBindingKeyBuilder  = store.NewKeyBuilder(roleBindingsPathPrefix)
)

func getRoleBindingPath(roleBinding *types.RoleBinding) string {
	return roleBindingKeyBuilder.WithResource(roleBinding).Build(roleBinding.Name)
}

func getRoleBindingsPath(ctx context.Context, name string) string {
	return roleBindingKeyBuilder.WithContext(ctx).Build(name)
}

// CreateRoleBinding ...
func (s *Store) CreateRoleBinding(ctx context.Context, roleBinding *types.RoleBinding) error {
	if err := roleBinding.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return s.create(ctx, getRoleBindingPath(roleBinding), roleBinding.Namespace, roleBinding)
}

// CreateOrUpdateRoleBinding ...
func (s *Store) CreateOrUpdateRoleBinding(ctx context.Context, roleBinding *types.RoleBinding) error {
	if err := roleBinding.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return s.createOrUpdate(ctx, getRoleBindingPath(roleBinding), roleBinding.Namespace, roleBinding)
}

// DeleteRoleBinding ...
func (s *Store) DeleteRoleBinding(ctx context.Context, name string) error {
	return s.delete(ctx, getRoleBindingsPath(ctx, name))
}

// GetRoleBinding ...
func (s *Store) GetRoleBinding(ctx context.Context, name string) (*types.RoleBinding, error) {
	role := &types.RoleBinding{}
	err := s.get(ctx, getRoleBindingsPath(ctx, name), role)
	return role, err
}

// ListRolesBinding ...
func (s *Store) ListRolesBindings(ctx context.Context) ([]*types.RoleBinding, error) {
	roles := []*types.RoleBinding{}
	err := s.list(ctx, getRoleBindingsPath, &roles)
	return roles, err
}

// UpdateRoleBinding ...
func (s *Store) UpdateRoleBinding(ctx context.Context, roleBinding *types.RoleBinding) error {
	if err := roleBinding.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return s.update(ctx, getRoleBindingPath(roleBinding), roleBinding.Namespace, roleBinding)
}
