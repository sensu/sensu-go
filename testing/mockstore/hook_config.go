package mockstore

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// DeleteHookConfigByName ...
func (s *MockStore) DeleteHookConfigByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetHookConfigs ...
func (s *MockStore) GetHookConfigs(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.HookConfig, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*corev2.HookConfig), args.Error(1)
}

// GetHookConfigByName ...
func (s *MockStore) GetHookConfigByName(ctx context.Context, name string) (*corev2.HookConfig, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*corev2.HookConfig), args.Error(1)
}

// UpdateHookConfig ...
func (s *MockStore) UpdateHookConfig(ctx context.Context, hook *corev2.HookConfig) error {
	args := s.Called(ctx, hook)
	return args.Error(0)
}
