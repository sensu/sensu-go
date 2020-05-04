package etcd

import (
	"context"
	"errors"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

const (
	hooksPathPrefix = "hooks"
)

var (
	hookKeyBuilder = store.NewKeyBuilder(hooksPathPrefix)
)

func getHookConfigPath(hook *corev2.HookConfig) string {
	return hookKeyBuilder.WithResource(hook).Build(hook.Name)
}

// GetHookConfigsPath gets the path of the hook config store.
func GetHookConfigsPath(ctx context.Context, name string) string {
	return hookKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteHookConfigByName deletes a HookConfig by name.
func (s *Store) DeleteHookConfigByName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name")
	}

	err := Delete(ctx, s.client, GetHookConfigsPath(ctx, name))
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			err = nil
		}
	}
	return err
}

// GetHookConfigs returns hook configurations for a namespace.
func (s *Store) GetHookConfigs(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.HookConfig, error) {
	hooks := []*corev2.HookConfig{}
	err := List(ctx, s.client, GetHookConfigsPath, &hooks, pred)
	return hooks, err
}

// GetHookConfigByName gets a HookConfig by name.
func (s *Store) GetHookConfigByName(ctx context.Context, name string) (*corev2.HookConfig, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	var hook corev2.HookConfig
	if err := Get(ctx, s.client, GetHookConfigsPath(ctx, name), &hook); err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			err = nil
		}
		return nil, err
	}

	return &hook, nil
}

// UpdateHookConfig updates a HookConfig.
func (s *Store) UpdateHookConfig(ctx context.Context, hook *corev2.HookConfig) error {
	if err := hook.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	return CreateOrUpdate(ctx, s.client, getHookConfigPath(hook), hook.Namespace, hook)
}
