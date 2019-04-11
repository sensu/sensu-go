package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// DeleteCheckConfigByName ...
func (s *MockStore) DeleteCheckConfigByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetCheckConfigs ...
func (s *MockStore) GetCheckConfigs(ctx context.Context, pred *store.SelectionPredicate) ([]*types.CheckConfig, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*types.CheckConfig), args.Error(1)
}

// GetCheckConfigByName ...
func (s *MockStore) GetCheckConfigByName(ctx context.Context, name string) (*types.CheckConfig, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*types.CheckConfig), args.Error(1)
}

// UpdateCheckConfig ...
func (s *MockStore) UpdateCheckConfig(ctx context.Context, check *types.CheckConfig) error {
	args := s.Called(ctx, check)
	return args.Error(0)
}

// GetCheckConfigWatcher ...
func (s *MockStore) GetCheckConfigWatcher(ctx context.Context) <-chan store.WatchEventCheckConfig {
	args := s.Called(ctx)
	return args.Get(0).(<-chan store.WatchEventCheckConfig)
}
