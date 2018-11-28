package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	checksPathPrefix = "checks"
)

var (
	checkKeyBuilder = store.NewKeyBuilder(checksPathPrefix)
)

func getCheckConfigPath(check *types.CheckConfig) string {
	return checkKeyBuilder.WithResource(check).Build(check.Name)
}

func getCheckConfigsPath(ctx context.Context, name string) string {
	return checkKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteCheckConfigByName deletes a CheckConfig by name.
func (s *Store) DeleteCheckConfigByName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name")
	}

	_, err := s.client.Delete(ctx, getCheckConfigsPath(ctx, name))
	return err
}

// GetCheckConfigs returns check configurations for an (optional) namespace.
func (s *Store) GetCheckConfigs(ctx context.Context) ([]*types.CheckConfig, error) {
	resp, err := query(ctx, s, getCheckConfigsPath)
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
		if check.Labels == nil {
			check.Labels = make(map[string]string)
		}
		if check.Annotations == nil {
			check.Annotations = make(map[string]string)
		}
		checksArray[i] = check
	}

	return checksArray, nil
}

// GetCheckConfigByName gets a CheckConfig by name.
func (s *Store) GetCheckConfigByName(ctx context.Context, name string) (*types.CheckConfig, error) {
	if name == "" {
		return nil, errors.New("must specify name")
	}

	resp, err := s.client.Get(ctx, getCheckConfigsPath(ctx, name))
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
	if check.Labels == nil {
		check.Labels = make(map[string]string)
	}
	if check.Annotations == nil {
		check.Annotations = make(map[string]string)
	}

	return check, nil
}

// UpdateCheckConfig updates a CheckConfig.
func (s *Store) UpdateCheckConfig(ctx context.Context, check *types.CheckConfig) error {
	if err := check.Validate(); err != nil {
		return err
	}

	checkBytes, err := json.Marshal(check)
	if err != nil {
		return err
	}

	cmp := clientv3.Compare(clientv3.Version(getNamespacePath(check.Namespace)), ">", 0)
	req := clientv3.OpPut(getCheckConfigPath(check), string(checkBytes))
	res, err := s.client.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the check %s in namespace %s",
			check.Name,
			check.Namespace,
		)
	}

	return nil
}
