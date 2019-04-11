package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// DeleteHandlerByName ...
func (s *MockStore) DeleteHandlerByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetHandlers ...
func (s *MockStore) GetHandlers(ctx context.Context, pred *store.SelectionPredicate) ([]*types.Handler, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*types.Handler), args.Error(1)
}

// GetHandlerByName ...
func (s *MockStore) GetHandlerByName(ctx context.Context, name string) (*types.Handler, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*types.Handler), args.Error(1)
}

// UpdateHandler ...
func (s *MockStore) UpdateHandler(ctx context.Context, handler *types.Handler) error {
	args := s.Called(handler)
	return args.Error(0)
}
