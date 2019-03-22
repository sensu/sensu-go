package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// DeleteEventFilterByName ...
func (s *MockStore) DeleteEventFilterByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetEventFilters ...
func (s *MockStore) GetEventFilters(ctx context.Context, pageSize int64, continueToken string) ([]*types.EventFilter, string, error) {
	args := s.Called(ctx, pageSize, continueToken)
	return args.Get(0).([]*types.EventFilter), args.String(1), args.Error(2)
}

// GetEventFilterByName ...
func (s *MockStore) GetEventFilterByName(ctx context.Context, name string) (*types.EventFilter, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*types.EventFilter), args.Error(1)
}

// UpdateEventFilter ...
func (s *MockStore) UpdateEventFilter(ctx context.Context, filter *types.EventFilter) error {
	args := s.Called(filter)
	return args.Error(0)
}
