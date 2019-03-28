package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// DeleteHookConfigByName ...
func (s *MockStore) DeleteHookConfigByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetHookConfigs ...
func (s *MockStore) GetHookConfigs(ctx context.Context, pageSize int64, continueToken string) ([]*types.HookConfig, string, error) {
	args := s.Called(ctx, pageSize, continueToken)
	return args.Get(0).([]*types.HookConfig), args.String(1), args.Error(2)
}

// GetHookConfigByName ...
func (s *MockStore) GetHookConfigByName(ctx context.Context, name string) (*types.HookConfig, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*types.HookConfig), args.Error(1)
}

// UpdateHookConfig ...
func (s *MockStore) UpdateHookConfig(ctx context.Context, hook *types.HookConfig) error {
	args := s.Called(ctx, hook)
	return args.Error(0)
}
