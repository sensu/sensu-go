package etcd

import (
	"context"
	"encoding/json"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sensu/sensu-go/types"
)

const (
	rolePathPrefix = "roles"
)

func getRolePath(name string) string {
	return path.Join(EtcdRoot, rolePathPrefix, name)
}

// GetRoles ...
func (s *Store) GetRoles(ctx context.Context) ([]*types.Role, error) {
	resp, err := s.kvc.Get(ctx, getRolePath(""), clientv3.WithPrefix())
	if err != nil {
		return []*types.Role{}, err
	}

	return unmarshalRole(resp.Kvs)
}

// GetRoleByName ...
func (s *Store) GetRoleByName(ctx context.Context, name string) (*types.Role, error) {
	resp, err := s.kvc.Get(ctx, getRolePath(name), clientv3.WithLimit(1))
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

// UpdateRole ...
func (s *Store) UpdateRole(ctx context.Context, role *types.Role) error {
	if err := role.Validate(); err != nil {
		return err
	}

	roleBytes, err := json.Marshal(role)
	if err != nil {
		return err
	}

	_, err = s.kvc.Put(ctx, getRolePath(role.Name), string(roleBytes))
	return err
}

// DeleteRoleByName ...
func (s *Store) DeleteRoleByName(ctx context.Context, name string) error {
	_, err := s.kvc.Delete(ctx, getRolePath(name))
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
