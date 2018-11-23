package etcd

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var (
	rolesPathPrefix = "rbac/roles"
	roleKeyBuilder  = store.NewKeyBuilder(rolesPathPrefix)
)

func getRolePath(role *types.Role) string {
	return roleKeyBuilder.WithResource(role).Build(role.Name)
}

func getRolesPath(ctx context.Context, name string) string {
	return roleKeyBuilder.WithContext(ctx).Build(name)
}

// CreateRole ...
func (s *Store) CreateRole(ctx context.Context, role *types.Role) error {
	if err := role.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return s.create(ctx, getRolePath(role), role.Namespace, role)
}

// CreateOrUpdateRole ...
func (s *Store) CreateOrUpdateRole(ctx context.Context, role *types.Role) error {
	if err := role.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return s.createOrUpdate(ctx, getRolePath(role), role.Namespace, role)
}

// DeleteRole ...
func (s *Store) DeleteRole(ctx context.Context, name string) error {
	return s.delete(ctx, getRolesPath(ctx, name))
}

// GetRole ...
func (s *Store) GetRole(ctx context.Context, name string) (*types.Role, error) {
	role := &types.Role{}
	err := s.get(ctx, getRolesPath(ctx, name), role)
	return role, err
}

// ListRoles ...
func (s *Store) ListRoles(ctx context.Context) ([]*types.Role, error) {
	roles := []*types.Role{}
	err := s.list(ctx, getRolesPath, &roles)
	return roles, err
}

// UpdateRole ...
func (s *Store) UpdateRole(ctx context.Context, role *types.Role) error {
	if err := role.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return s.update(ctx, getRolePath(role), role.Namespace, role)
}
