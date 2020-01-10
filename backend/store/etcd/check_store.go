package etcd

import (
	"context"
	"errors"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
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

// GetCheckConfigsPath gets the path of the check config store.
func GetCheckConfigsPath(ctx context.Context, name string) string {
	return checkKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteCheckConfigByName deletes a CheckConfig by name.
func (s *Store) DeleteCheckConfigByName(ctx context.Context, name string) error {
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	if _, err := s.client.Delete(ctx, GetCheckConfigsPath(ctx, name)); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	return nil
}

// GetCheckConfigs returns check configurations for an (optional) namespace.
func (s *Store) GetCheckConfigs(ctx context.Context, pred *store.SelectionPredicate) ([]*types.CheckConfig, error) {
	checks := []*types.CheckConfig{}
	err := List(ctx, s.client, GetCheckConfigsPath, &checks, pred)
	return checks, err
}

// GetCheckConfigByName gets a CheckConfig by name.
func (s *Store) GetCheckConfigByName(ctx context.Context, name string) (*types.CheckConfig, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	resp, err := s.client.Get(ctx, GetCheckConfigsPath(ctx, name))
	if err != nil {
		return nil, &store.ErrInternal{Message: err.Error()}
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	checkBytes := resp.Kvs[0].Value
	check := &types.CheckConfig{}
	if err := unmarshal(checkBytes, check); err != nil {
		return nil, &store.ErrDecode{Err: err}
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
		return &store.ErrNotValid{Err: err}
	}

	checkBytes, err := proto.Marshal(check)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}

	cmp := clientv3.Compare(clientv3.Version(getNamespacePath(check.Namespace)), ">", 0)
	req := clientv3.OpPut(getCheckConfigPath(check), string(checkBytes))
	res, err := s.client.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	if !res.Succeeded {
		return &store.ErrNamespaceMissing{Namespace: check.Namespace}
	}

	return nil
}
