package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sensu/sensu-go/types"
)

const (
	organizationsPathPrefix = "organizations"
)

func getOrganizationsPath(name string) string {
	return path.Join(etcdRoot, organizationsPathPrefix, name)
}

// DeleteOrganizationByName deletes the organization named *name*
func (s *etcdStore) DeleteOrganizationByName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name")
	}

	resp, err := s.kvc.Delete(context.TODO(), getOrganizationsPath(name), clientv3.WithPrefix())
	if err != nil {
		return err
	}

	if resp.Deleted != 1 {
		return fmt.Errorf("organization %s does not exist", name)
	}

	return nil
}

// GetOrganizationByName returns a single organization named *name*
func (s *etcdStore) GetOrganizationByName(ctx context.Context, name string) (*types.Organization, error) {
	resp, err := s.kvc.Get(
		context.TODO(),
		getOrganizationsPath(name),
		clientv3.WithLimit(1),
	)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) != 1 {
		return nil, fmt.Errorf("organization %s does not exist", name)
	}

	orgs, err := unmarshalOrganizations(resp.Kvs)
	if err != nil {
		return &types.Organization{}, err
	}

	return orgs[0], nil
}

// GetOrganizations returns all organizations
func (s *etcdStore) GetOrganizations(ctx context.Context) ([]*types.Organization, error) {
	resp, err := s.kvc.Get(
		context.TODO(),
		getOrganizationsPath(""),
		clientv3.WithPrefix(),
	)

	if err != nil {
		return []*types.Organization{}, err
	}

	return unmarshalOrganizations(resp.Kvs)
}

// UpdateOrganization updates an organization with the provided org
func (s *etcdStore) UpdateOrganization(ctx context.Context, org *types.Organization) error {
	if err := org.Validate(); err != nil {
		return err
	}

	bytes, err := json.Marshal(org)
	if err != nil {
		return err
	}

	_, err = s.kvc.Put(
		context.TODO(),
		getOrganizationsPath(org.Name),
		string(bytes),
	)

	if err != nil {
		return err
	}

	return nil
}

func unmarshalOrganizations(kvs []*mvccpb.KeyValue) ([]*types.Organization, error) {
	s := make([]*types.Organization, len(kvs))
	for i, kv := range kvs {
		org := &types.Organization{}
		s[i] = org
		if err := json.Unmarshal(kv.Value, org); err != nil {
			return nil, err
		}
	}

	return s, nil
}
