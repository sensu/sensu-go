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

// GetRolesPath gets the path of the role store.
func GetRolesPath(ctx context.Context, name string) string {
	return roleKeyBuilder.WithContext(ctx).Build(name)
}

// CreateRole ...
func (s *Store) CreateRole(ctx context.Context, role *types.Role) error {
	if err := role.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return Create(ctx, s.client, getRolePath(role), role.Namespace, role)
}

// CreateOrUpdateRole ...
func (s *Store) CreateOrUpdateRole(ctx context.Context, role *types.Role) error {
	if err := role.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return CreateOrUpdate(ctx, s.client, getRolePath(role), role.Namespace, role)
}

// DeleteRole ...
func (s *Store) DeleteRole(ctx context.Context, name string) error {
	return Delete(ctx, s.client, GetRolesPath(ctx, name))
}

// GetRole ...
func (s *Store) GetRole(ctx context.Context, name string) (*types.Role, error) {
	role := &types.Role{}
	err := Get(ctx, s.client, GetRolesPath(ctx, name), role)
	return role, err
}

// ListRoles ...
func (s *Store) ListRoles(ctx context.Context, pred *store.SelectionPredicate) ([]*types.Role, error) {
	if pred == nil {
		pred = &store.SelectionPredicate{}
	}
	roles := []*types.Role{}
	err := List(ctx, s.client, GetRolesPath, &roles, pred)
	return roles, err
}

// UpdateRole ...
func (s *Store) UpdateRole(ctx context.Context, role *types.Role) error {
	if err := role.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return Update(ctx, s.client, getRolePath(role), role.Namespace, role)
}
