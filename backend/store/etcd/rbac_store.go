package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sensu/sensu-go/types"
)

const (
	rolePathPrefix = "roles"
)

func getRolePath(org, name string) string {
	return path.Join(etcdRoot, rolePathPrefix, org, name)
}

// GetRoles ...
func (s *etcdStore) GetRoles(org string) ([]*types.Role, error) {
	resp, err := s.kvc.Get(
		context.TODO(),
		getRolePath(org, ""),
		clientv3.WithPrefix(),
	)

	if err != nil {
		return []*types.Role{}, err
	}

	return unmarshalRole(resp.Kvs)
}

// GetRole ...
func (s *etcdStore) GetRole(org, name string) (*types.Role, error) {
	resp, err := s.kvc.Get(
		context.TODO(),
		getRolePath(org, name),
		clientv3.WithPrefix(),
	)

	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return &types.Role{}, nil
	}

	roles, err := unmarshalRole(resp.Kvs)
	if err != nil {
		return []*types.Role{}, err
	}

	return roles[0], nil
}

// CreateRole ...
func (s *etcdStore) CreateRole(role *types.Role) error {
	if err := role.Validate(); err != nil {
		return err
	}

	roleBytes, err := json.Marshal(role)
	if err != nil {
		return err
	}

	_, err = s.kvc.Put(
		context.TODO(),
		getRolePath(role.Organization, role.Name),
		string(roleBytes),
	)

	if err != nil {
		return err
	}

	return nil
}

// UpdateRole ...
func (s *etcdStore) UpdateRole(role *types.Role) error {
	if err := role.Validate(); err != nil {
		return err
	}

	roleBytes, err := json.Marshal(role)
	if err != nil {
		return err
	}

	_, err = s.kvc.Put(
		context.TODO(),
		getRolePath(role.Organization, role.Name),
		string(roleBytes),
	)

	if err != nil {
		return err
	}

	return nil
}

// DeleteRoleByName ...
func (s *etcdStore) DeleteRoleByName(org, name string) error {
	if org == "" || name == "" {
		return errors.New("must specify organization and name")
	}
	_, err := s.kvc.Delete(context.TODO(), getRolePath(org, name))
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
