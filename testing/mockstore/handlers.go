package mockstore

import (
	"context"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// DeleteHandlerByName ...
func (s *MockStore) DeleteHandlerByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetHandlers ...
func (s *MockStore) GetHandlers(ctx context.Context, pred *store.SelectionPredicate) ([]*v2.Handler, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*v2.Handler), args.Error(1)
}

// GetHandlerByName ...
func (s *MockStore) GetHandlerByName(ctx context.Context, name string) (*v2.Handler, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*v2.Handler), args.Error(1)
}

// UpdateHandler ...
func (s *MockStore) UpdateHandler(ctx context.Context, handler *v2.Handler) error {
	args := s.Called(handler)
	return args.Error(0)
}
