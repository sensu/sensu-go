package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// DeleteFilterByName ...
func (s *MockStore) DeleteFilterByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetFilters ...
func (s *MockStore) GetFilters(ctx context.Context) ([]*types.Filter, error) {
	args := s.Called(ctx)
	return args.Get(0).([]*types.Filter), args.Error(1)
}

// GetFilterByName ...
func (s *MockStore) GetFilterByName(ctx context.Context, name string) (*types.Filter, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*types.Filter), args.Error(1)
}

// UpdateFilter ...
func (s *MockStore) UpdateFilter(ctx context.Context, filter *types.Filter) error {
	args := s.Called(filter)
	return args.Error(0)
}
