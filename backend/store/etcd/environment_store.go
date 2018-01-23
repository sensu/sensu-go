package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var (
	environmentsPathPrefix = "environments"
	environmentKeyBuilder  = store.NewKeyBuilder(environmentsPathPrefix)
)

func getEnvironmentsPath(org, env string) string {
	return environmentKeyBuilder.WithOrg(org).Build(env)
}

// DeleteEnvironment deletes an environment
func (s *etcdStore) DeleteEnvironment(ctx context.Context, env *types.Environment) error {
	if err := env.Validate(); err != nil {
		return err
	}

	org := env.Organization

	ctx = context.WithValue(ctx, types.OrganizationKey, org)
	ctx = context.WithValue(ctx, types.EnvironmentKey, env.Name)

	// Validate whether there are any resources referencing the organization
	getresp, err := s.kvc.Txn(ctx).Then(
		v3.OpGet(checkKeyBuilder.WithContext(ctx).Build(), v3.WithPrefix(), v3.WithCountOnly()),
		v3.OpGet(entityKeyBuilder.WithContext(ctx).Build(), v3.WithPrefix(), v3.WithCountOnly()),
		v3.OpGet(assetKeyBuilder.WithContext(ctx).Build(), v3.WithPrefix(), v3.WithCountOnly()),
		v3.OpGet(handlerKeyBuilder.WithContext(ctx).Build(), v3.WithPrefix(), v3.WithCountOnly()),
		v3.OpGet(mutatorKeyBuilder.WithContext(ctx).Build(), v3.WithPrefix(), v3.WithCountOnly()),
	).Commit()
	if err != nil {
		return err
	}
	for _, r := range getresp.Responses {
		if r.GetResponseRange().Count > 0 {
			return errors.New("environment is not empty") // TODO
		}
	}

	// Validate that there are no roles referencing the organization
	roles, err := s.GetRoles(ctx)
	if err != nil {
		return err
	}
	for _, role := range roles {
		for _, rule := range role.Rules {
			if rule.Organization == org && rule.Environment == env.Name {
				return fmt.Errorf("environment is not empty; role '%s' references it", role.Name)
			}
		}
	}

	resp, err := s.kvc.Delete(ctx, getEnvironmentsPath(org, env.Name), v3.WithPrefix())
	if err != nil {
		return err
	}

	if resp.Deleted != 1 {
		return fmt.Errorf("environment %s/%s does not exist", org, env.Name)
	}

	return nil
}

// GetEnvironment returns a single environment
func (s *etcdStore) GetEnvironment(ctx context.Context, org, env string) (*types.Environment, error) {
	resp, err := s.kvc.Get(
		ctx,
		getEnvironmentsPath(org, env),
		v3.WithLimit(1),
	)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) != 1 {
		// DNE, but not an error
		return nil, nil
	}

	envs, err := unmarshalEnvironments(resp.Kvs)
	if err != nil {
		return &types.Environment{}, err
	}

	return envs[0], nil
}

// GetOrganizations returns all organizations
func (s *etcdStore) GetEnvironments(ctx context.Context, org string) ([]*types.Environment, error) {
	// Support "*" as a wildcard
	if org == "*" {
		org = ""
	}

	resp, err := s.kvc.Get(ctx, getEnvironmentsPath(org, ""), v3.WithPrefix())

	if err != nil {
		return []*types.Environment{}, err
	}

	return unmarshalEnvironments(resp.Kvs)
}

// UpdateEnvironment updates an environment
func (s *etcdStore) UpdateEnvironment(ctx context.Context, env *types.Environment) error {
	if err := env.Validate(); err != nil {
		return err
	}

	bytes, err := json.Marshal(env)
	if err != nil {
		return err
	}

	org := env.Organization

	// We need to prepare a transaction to verify that the organization under
	// which we are creating this environment exists
	cmp := v3.Compare(v3.Version(getOrganizationsPath(org)), ">", 0)
	req := v3.OpPut(getEnvironmentsPath(org, env.Name), string(bytes))
	res, err := s.kvc.Txn(ctx).If(cmp).Then(req).Commit()
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
