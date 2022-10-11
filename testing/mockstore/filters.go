package mockstore

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// DeleteEventFilterByName ...
func (s *MockStore) DeleteEventFilterByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetEventFilters ...
func (s *MockStore) GetEventFilters(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.EventFilter, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*corev2.EventFilter), args.Error(1)
}

// GetEventFilterByName ...
func (s *MockStore) GetEventFilterByName(ctx context.Context, name string) (*corev2.EventFilter, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*corev2.EventFilter), args.Error(1)
}

// UpdateEventFilter ...
func (s *MockStore) UpdateEventFilter(ctx context.Context, filter *corev2.EventFilter) error {
	args := s.Called(filter)
	return args.Error(0)
}
