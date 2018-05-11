package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sensu/sensu-go/types"
)

const (
	organizationsPathPrefix = "organizations"
)

func getOrganizationsPath(name string) string {
	return path.Join(EtcdRoot, organizationsPathPrefix, name)
}

// CreateOrganization creates an organization and a default environment with the
// provided org
func (s *Store) CreateOrganization(ctx context.Context, org *types.Organization) error {
	if err := org.Validate(); err != nil {
		return err
	}

	orgBytes, err := json.Marshal(org)
	if err != nil {
		return err
	}

	// Define a default environment for this new organization
	env := types.Environment{
		Name:         "default",
		Description:  "Default environment",
		Organization: org.Name,
	}

	envBytes, err := json.Marshal(env)
	if err != nil {
		return err
	}

	orgKey := getOrganizationsPath(org.Name)

	res, err := s.client.Txn(ctx).
		If(
			// Ensure the organization does not already exist
			v3.Compare(v3.Version(orgKey), "=", 0)).
		Then(
			// Create both the organization and a default environment
			v3.OpPut(orgKey, string(orgBytes)),
			v3.OpPut(getEnvironmentsPath(org.Name, env.Name), string(envBytes)),
		).Commit()
	if err != nil {
		return err
	}

	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the organization %s with a default environment",
			org.Name,
		)
	}

	return err
}

// DeleteOrganizationByName deletes the organization named *name*
func (s *Store) DeleteOrganizationByName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name")
	}

	// Validate whether there are any resources referencing the organization
	getresp, err := s.client.Txn(ctx).Then(
		v3.OpGet(checkKeyBuilder.WithOrg(name).Build(), v3.WithPrefix(), v3.WithCountOnly()),
		v3.OpGet(entityKeyBuilder.WithOrg(name).Build(), v3.WithPrefix(), v3.WithCountOnly()),
		v3.OpGet(assetKeyBuilder.WithOrg(name).Build(), v3.WithPrefix(), v3.WithCountOnly()),
		v3.OpGet(handlerKeyBuilder.WithOrg(name).Build(), v3.WithPrefix(), v3.WithCountOnly()),
		v3.OpGet(mutatorKeyBuilder.WithOrg(name).Build(), v3.WithPrefix(), v3.WithCountOnly()),
		v3.OpGet(environmentKeyBuilder.WithOrg(name).Build(), v3.WithPrefix(), v3.WithCountOnly()),
	).Commit()
	if err != nil {
		return err
	}
	for _, r := range getresp.Responses {
		if r.GetResponseRange().Count > 0 {
			return errors.New("organization is not empty") // TODO
		}
	}

	// Validate that there are no roles referencing the organization
	roles, err := s.GetRoles(ctx)
	if err != nil {
		return err
	}
	for _, role := range roles {
		for _, rule := range role.Rules {
			if rule.Organization == name {
				return fmt.Errorf("organization is not empty; role '%s' references it", role.Name)
			}
		}
	}

	// Delete the resource
	resp, err := s.client.Delete(ctx, getOrganizationsPath(name), v3.WithPrefix())
	if err != nil {
		return err
	}

	if resp.Deleted != 1 {
		return fmt.Errorf("organization %s does not exist", name)
	}

	return nil
}

// GetOrganizationByName returns a single organization named *name*
func (s *Store) GetOrganizationByName(ctx context.Context, name string) (*types.Organization, error) {
	resp, err := s.client.Get(
		ctx,
		getOrganizationsPath(name),
		v3.WithLimit(1),
	)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	orgs, err := unmarshalOrganizations(resp.Kvs)
	if err != nil {
		return &types.Organization{}, err
	}

	return orgs[0], nil
}

// GetOrganizations returns all organizations
func (s *Store) GetOrganizations(ctx context.Context) ([]*types.Organization, error) {
	resp, err := s.client.Get(
		ctx,
		getOrganizationsPath(""),
		v3.WithPrefix(),
	)

	if err != nil {
		return []*types.Organization{}, err
	}

	return unmarshalOrganizations(resp.Kvs)
}

// UpdateOrganization updates an organization with the provided org
func (s *Store) UpdateOrganization(ctx context.Context, org *types.Organization) error {
	if err := org.Validate(); err != nil {
		return err
	}

	bytes, err := json.Marshal(org)
	if err != nil {
		return err
	}

	_, err = s.client.Put(ctx, getOrganizationsPath(org.Name), string(bytes))

	return err
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
