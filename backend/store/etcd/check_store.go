package etcd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

func getCheckConfigsPath(name string) string {
	return fmt.Sprintf("%s/checks/%s", etcdRoot, name)
}

func (s *etcdStore) GetCheckConfigs() ([]*types.CheckConfig, error) {
	resp, err := s.kvc.Get(context.TODO(), getCheckConfigsPath(""), clientv3.WithPrefix())
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

func (s *etcdStore) GetCheckConfigByName(name string) (*types.CheckConfig, error) {
	resp, err := s.kvc.Get(context.TODO(), getCheckConfigsPath(name))
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

func (s *etcdStore) DeleteCheckConfigByName(name string) error {
	_, err := s.kvc.Delete(context.TODO(), getCheckConfigsPath(name))
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

	_, err = s.kvc.Put(context.TODO(), getCheckConfigsPath(check.Name), string(checkBytes))
	if err != nil {
		return err
	}

	return nil
}
