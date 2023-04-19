package mockstore

import (
	"context"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// DeleteHookConfigByName ...
func (s *MockStore) DeleteHookConfigByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetHookConfigs ...
func (s *MockStore) GetHookConfigs(ctx context.Context, pred *store.SelectionPredicate) ([]*v2.HookConfig, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*v2.HookConfig), args.Error(1)
}

// GetHookConfigByName ...
func (s *MockStore) GetHookConfigByName(ctx context.Context, name string) (*v2.HookConfig, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*v2.HookConfig), args.Error(1)
}

// UpdateHookConfig ...
func (s *MockStore) UpdateHookConfig(ctx context.Context, hook *v2.HookConfig) error {
	args := s.Called(ctx, hook)
	return args.Error(0)
}
