package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

const (
	hooksPathPrefix = "hooks"
)

var (
	hookKeyBuilder = store.NewKeyBuilder(hooksPathPrefix)
)

func getHookConfigPath(hook *types.HookConfig) string {
	return hookKeyBuilder.WithResource(hook).Build(hook.Name)
}

func getHookConfigsPath(ctx context.Context, name string) string {
	return hookKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteHookConfigByName deletes a HookConfig by name.
func (s *Store) DeleteHookConfigByName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name")
	}

	_, err := s.client.Delete(ctx, getHookConfigsPath(ctx, name))
	return err
}

// GetHookConfigs returns hook configurations for a namespace.
func (s *Store) GetHookConfigs(ctx context.Context, pageSize int64, continueToken string) (hooks []*corev2.HookConfig, newContinueToken string, err error) {
	opts := []clientv3.OpOption{
		clientv3.WithLimit(pageSize),
	}

	keyPrefix := getHookConfigsPath(ctx, "")
	rangeEnd := clientv3.GetPrefixRangeEnd(keyPrefix)
	opts = append(opts, clientv3.WithRange(rangeEnd))

	resp, err := s.client.Get(ctx, path.Join(keyPrefix, continueToken), opts...)
	if err != nil {
		return nil, "", err
	}
	if len(resp.Kvs) == 0 {
		return []*corev2.HookConfig{}, "", nil
	}

	for _, kv := range resp.Kvs {
		hook := &corev2.HookConfig{}
		err = json.Unmarshal(kv.Value, hook)
		if err != nil {
			return nil, "", err
		}

		hooks = append(hooks, hook)
	}

	if pageSize != 0 && resp.Count > pageSize {
		lastHook := hooks[len(hooks)-1]
		newContinueToken = computeContinueToken(ctx, lastHook)
	}

	return hooks, newContinueToken, nil
}

// GetHookConfigByName gets a HookConfig by name.
func (s *Store) GetHookConfigByName(ctx context.Context, name string) (*types.HookConfig, error) {
	if name == "" {
		return nil, errors.New("must specify name")
	}

	resp, err := s.client.Get(ctx, getHookConfigsPath(ctx, name))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	hookBytes := resp.Kvs[0].Value
	hook := &types.HookConfig{}
	if err := json.Unmarshal(hookBytes, hook); err != nil {
		return nil, err
	}

	return hook, nil
}

// UpdateHookConfig updates a HookConfig.
func (s *Store) UpdateHookConfig(ctx context.Context, hook *types.HookConfig) error {
	if err := hook.Validate(); err != nil {
		return err
	}

	hookBytes, err := json.Marshal(hook)
	if err != nil {
		return err
	}

	cmp := clientv3.Compare(clientv3.Version(getNamespacePath(hook.Namespace)), ">", 0)
	req := clientv3.OpPut(getHookConfigPath(hook), string(hookBytes))
	res, err := s.client.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the hook %s in namespace %s",
			hook.Name,
			hook.Namespace,
		)
	}

	return nil
}
