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

// GetRoleBindingsPath gets the path of the role binding store.
func GetRoleBindingsPath(ctx context.Context, name string) string {
	return roleBindingKeyBuilder.WithContext(ctx).Build(name)
}

// CreateRoleBinding ...
func (s *Store) CreateRoleBinding(ctx context.Context, roleBinding *types.RoleBinding) error {
	if err := roleBinding.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return Create(ctx, s.client, getRoleBindingPath(roleBinding), roleBinding.Namespace, roleBinding)
}

// CreateOrUpdateRoleBinding ...
func (s *Store) CreateOrUpdateRoleBinding(ctx context.Context, roleBinding *types.RoleBinding) error {
	if err := roleBinding.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return CreateOrUpdate(ctx, s.client, getRoleBindingPath(roleBinding), roleBinding.Namespace, roleBinding)
}

// DeleteRoleBinding ...
func (s *Store) DeleteRoleBinding(ctx context.Context, name string) error {
	return Delete(ctx, s.client, GetRoleBindingsPath(ctx, name))
}

// GetRoleBinding ...
func (s *Store) GetRoleBinding(ctx context.Context, name string) (*types.RoleBinding, error) {
	role := &types.RoleBinding{}
	err := Get(ctx, s.client, GetRoleBindingsPath(ctx, name), role)
	return role, err
}

// ListRoleBindings ...
func (s *Store) ListRoleBindings(ctx context.Context, pred *store.SelectionPredicate) ([]*types.RoleBinding, error) {
	roles := []*types.RoleBinding{}
	err := List(ctx, s.client, GetRoleBindingsPath, &roles, pred)
	return roles, err
}

// UpdateRoleBinding ...
func (s *Store) UpdateRoleBinding(ctx context.Context, roleBinding *types.RoleBinding) error {
	if err := roleBinding.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return Update(ctx, s.client, getRoleBindingPath(roleBinding), roleBinding.Namespace, roleBinding)
}
