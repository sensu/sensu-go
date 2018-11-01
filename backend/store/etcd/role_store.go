package etcd

import (
	"context"
	"encoding/json"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
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

// DeleteRole ...
func (s *Store) DeleteRole(ctx context.Context, name string) error {
	_, err := s.client.Delete(ctx, getRolesPath(ctx, name))
	return err
}

func unmarshalRole(kvs []*mvccpb.KeyValue) ([]*types.Role, error) {
	rolesArray := make([]*types.Role, len(kvs))
	for i, kv := range kvs {
		role := &types.Role{}
		rolesArray[i] = role
		if err := json.Unmarshal(kv.Value, role); err != nil {
			return nil, err
		}
	}

	return rolesArray, nil
}

// GetRole ...
func (s *Store) GetRole(ctx context.Context, name string) (*types.Role, error) {
	resp, err := s.client.Get(ctx, getRolesPath(ctx, name), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	roles, err := unmarshalRole(resp.Kvs)
	if err != nil {
		return nil, err
	}

	return roles[0], nil
}

// ListRoles ...
func (s *Store) ListRoles(ctx context.Context) ([]*types.Role, error) {
	resp, err := query(ctx, s, getRolesPath)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return []*types.Role{}, nil
	}

	return unmarshalRole(resp.Kvs)
}

// UpdateRole ...
func (s *Store) UpdateRole(ctx context.Context, role *types.Role) error {
	if err := role.Validate(); err != nil {
		return err
	}

	roleBytes, err := json.Marshal(role)
	if err != nil {
		return err
	}

	_, err = s.client.Put(ctx, getRolePath(role), string(roleBytes))
	return err
}
