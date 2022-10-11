package etcd

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	corev2 "github.com/sensu/core/v2"
)

var (
	rolesPathPrefix = "rbac/roles"
	roleKeyBuilder  = store.NewKeyBuilder(rolesPathPrefix)
)

func getRolePath(role *corev2.Role) string {
	return roleKeyBuilder.WithResource(role).Build(role.Name)
}

// GetRolesPath gets the path of the role store.
func GetRolesPath(ctx context.Context, name string) string {
	return roleKeyBuilder.WithContext(ctx).Build(name)
}

// CreateRole ...
func (s *Store) CreateRole(ctx context.Context, role *corev2.Role) error {
	if err := role.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return Create(ctx, s.client, getRolePath(role), role.Namespace, role)
}

// CreateOrUpdateRole ...
func (s *Store) CreateOrUpdateRole(ctx context.Context, role *corev2.Role) error {
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
func (s *Store) GetRole(ctx context.Context, name string) (*corev2.Role, error) {
	role := &corev2.Role{}
	err := Get(ctx, s.client, GetRolesPath(ctx, name), role)
	return role, err
}

// ListRoles ...
func (s *Store) ListRoles(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Role, error) {
	roles := []*corev2.Role{}
	err := List(ctx, s.client, GetRolesPath, &roles, pred)
	return roles, err
}

// UpdateRole ...
func (s *Store) UpdateRole(ctx context.Context, role *corev2.Role) error {
	if err := role.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	return Update(ctx, s.client, getRolePath(role), role.Namespace, role)
}
