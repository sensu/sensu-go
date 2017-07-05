package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

const (
	checksPathPrefix = "checks"
)

func getCheckConfigsPath(org, name string) string {
	return path.Join(etcdRoot, checksPathPrefix, org, name)
}

// GetCheckConfigs returns check configurations for an (optional) organization.
// If org is the empty string, it returns all check configs.
func (s *etcdStore) GetCheckConfigs(org string) ([]*types.CheckConfig, error) {
	// Verify that the organization exist
	if org != "" {
		if _, err := s.GetOrganizationByName(org); err != nil {
			return nil, fmt.Errorf("the organization '%s' is invalid", org)
		}
	}

	resp, err := s.kvc.Get(context.TODO(), getCheckConfigsPath(org, ""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return []*types.CheckConfig{}, nil
	}

	checksArray := make([]*types.CheckConfig, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		check := &types.CheckConfig{}
		err = json.Unmarshal(kv.Value, check)
		if err != nil {
			return nil, err
		}
		checksArray[i] = check
	}

	return checksArray, nil
}

func (s *etcdStore) GetCheckConfigByName(org, name string) (*types.CheckConfig, error) {
	if org == "" || name == "" {
		return nil, errors.New("must specify organization and name")
	}

	// Verify that the organization exist
	if _, err := s.GetOrganizationByName(org); err != nil {
		return nil, fmt.Errorf("the organization '%s' is invalid", org)
	}

	resp, err := s.kvc.Get(context.TODO(), getCheckConfigsPath(org, name))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	checkBytes := resp.Kvs[0].Value
	check := &types.CheckConfig{}
	if err := json.Unmarshal(checkBytes, check); err != nil {
		return nil, err
	}

	return check, nil
}

func (s *etcdStore) DeleteCheckConfigByName(org, name string) error {
	if org == "" || name == "" {
		return errors.New("must specify organization and name")
	}

	// Verify that the organization exist
	if _, err := s.GetOrganizationByName(org); err != nil {
		return fmt.Errorf("the organization '%s' is invalid", org)
	}

	_, err := s.kvc.Delete(context.TODO(), getCheckConfigsPath(org, name))
	return err
}

func (s *etcdStore) UpdateCheckConfig(check *types.CheckConfig) error {
	if err := check.Validate(); err != nil {
		return err
	}

	checkBytes, err := json.Marshal(check)
	if err != nil {
		return err
	}

	_, err = s.kvc.Put(context.TODO(), getCheckConfigsPath(check.Organization, check.Name), string(checkBytes))
	if err != nil {
		return err
	}

	return nil
}
