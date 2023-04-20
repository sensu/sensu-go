package mockstore

import (
	"context"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// DeleteEventFilterByName ...
func (s *MockStore) DeleteEventFilterByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetEventFilters ...
func (s *MockStore) GetEventFilters(ctx context.Context, pred *store.SelectionPredicate) ([]*v2.EventFilter, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*v2.EventFilter), args.Error(1)
}

// GetEventFilterByName ...
func (s *MockStore) GetEventFilterByName(ctx context.Context, name string) (*v2.EventFilter, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*v2.EventFilter), args.Error(1)
}

// UpdateEventFilter ...
func (s *MockStore) UpdateEventFilter(ctx context.Context, filter *v2.EventFilter) error {
	args := s.Called(filter)
	return args.Error(0)
}
