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
	environmentsPathPrefix = "environments"
)

func getEnvironmentsPath(org, env string) string {
	orgPath := getOrganizationsPath(org)
	return path.Join(orgPath, environmentsPathPrefix, env)
}

// DeleteEnvironment deletes an environment
func (s *etcdStore) DeleteEnvironment(ctx context.Context, org, env string) error {
	if org == "" || env == "" {
		return errors.New("must specify organization and environment name")
	}

	resp, err := s.kvc.Delete(context.TODO(), getEnvironmentsPath(org, env), clientv3.WithPrefix())
	if err != nil {
		return err
	}

	if resp.Deleted != 1 {
		return fmt.Errorf("environment %s/%s does not exist", org, env)
	}

	return nil
}

// GetEnvironment returns a single environment
func (s *etcdStore) GetEnvironment(ctx context.Context, org, env string) (*types.Environment, error) {
	resp, err := s.kvc.Get(
		context.TODO(),
		getEnvironmentsPath(org, env),
		clientv3.WithLimit(1),
	)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) != 1 {
		return nil, fmt.Errorf("environment %s/%s does not exist", org, env)
	}

	envs, err := unmarshalEnvironments(resp.Kvs)
	if err != nil {
		return &types.Environment{}, err
	}

	return envs[0], nil
}

// GetOrganizations returns all organizations
func (s *etcdStore) GetEnvironments(ctx context.Context, org string) ([]*types.Environment, error) {
	resp, err := s.kvc.Get(
		context.TODO(),
		getEnvironmentsPath(org, "/"),
		clientv3.WithPrefix(),
	)

	if err != nil {
		return []*types.Environment{}, err
	}

	return unmarshalEnvironments(resp.Kvs)
}

// UpdateEnvironment updates an environment
func (s *etcdStore) UpdateEnvironment(ctx context.Context, org string, env *types.Environment) error {
	if err := env.Validate(); err != nil {
		return err
	}

	bytes, err := json.Marshal(env)
	if err != nil {
		return err
	}

	// We need to prepare a transaction to verify that the organization under
	// which we are creating this environment exists
	cmp := clientv3.Compare(clientv3.Version(getOrganizationsPath(org)), ">", 0)
	req := clientv3.OpPut(getEnvironmentsPath(org, env.Name), string(bytes))
	res, err := s.kvc.Txn(context.TODO()).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"the organization %s does not exist, cannot create the environment %s",
			org, env.Name,
		)
	}

	return nil
}

func unmarshalEnvironments(kvs []*mvccpb.KeyValue) ([]*types.Environment, error) {
	s := make([]*types.Environment, len(kvs))
	for i, kv := range kvs {
		env := &types.Environment{}
		s[i] = env
		if err := json.Unmarshal(kv.Value, env); err != nil {
			return nil, err
		}
	}

	return s, nil
}
