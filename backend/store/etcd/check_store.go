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

func getCheckConfigPath(check *types.CheckConfig) string {
	return path.Join(etcdRoot, checksPathPrefix, check.Organization, check.Environment, check.Name)
}

func getCheckConfigsPath(ctx context.Context, name string) string {
	env := environment(ctx)
	org := organization(ctx)

	return path.Join(etcdRoot, checksPathPrefix, org, env, name)
}

func (s *etcdStore) DeleteCheckConfigByName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name")
	}

	_, err := s.kvc.Delete(context.TODO(), getCheckConfigsPath(ctx, name))
	return err
}

// GetCheckConfigs returns check configurations for an (optional) organization.
// If org is the empty string, it returns all check configs.
func (s *etcdStore) GetCheckConfigs(ctx context.Context) ([]*types.CheckConfig, error) {
	resp, err := s.kvc.Get(context.TODO(), getCheckConfigsPath(ctx, ""), clientv3.WithPrefix())
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

func (s *etcdStore) GetCheckConfigByName(ctx context.Context, name string) (*types.CheckConfig, error) {
	if name == "" {
		return nil, errors.New("must specify name")
	}

	resp, err := s.kvc.Get(context.TODO(), getCheckConfigsPath(ctx, name))
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

func (s *etcdStore) UpdateCheckConfig(ctx context.Context, check *types.CheckConfig) error {
	if err := check.Validate(); err != nil {
		return err
	}

	checkBytes, err := json.Marshal(check)
	if err != nil {
		return err
	}

	cmp := clientv3.Compare(clientv3.Version(getEnvironmentsPath(check.Organization, check.Environment)), ">", 0)
	req := clientv3.OpPut(getCheckConfigPath(check), string(checkBytes))
	res, err := s.kvc.Txn(context.TODO()).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the check %s in environment %s/%s",
			check.Name,
			check.Organization,
			check.Environment,
		)
	}

	return nil
}
