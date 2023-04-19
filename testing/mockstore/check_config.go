package mockstore

import (
	"context"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// DeleteCheckConfigByName ...
func (s *MockStore) DeleteCheckConfigByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetCheckConfigs ...
func (s *MockStore) GetCheckConfigs(ctx context.Context, pred *store.SelectionPredicate) ([]*v2.CheckConfig, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*v2.CheckConfig), args.Error(1)
}

// GetCheckConfigByName ...
func (s *MockStore) GetCheckConfigByName(ctx context.Context, name string) (*v2.CheckConfig, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*v2.CheckConfig), args.Error(1)
}

// UpdateCheckConfig ...
func (s *MockStore) UpdateCheckConfig(ctx context.Context, check *v2.CheckConfig) error {
	args := s.Called(ctx, check)
	return args.Error(0)
}

// GetCheckConfigWatcher ...
func (s *MockStore) GetCheckConfigWatcher(ctx context.Context) <-chan store.WatchEventCheckConfig {
	args := s.Called(ctx)
	return args.Get(0).(<-chan store.WatchEventCheckConfig)
}
