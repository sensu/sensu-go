package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// DeleteEntity ...
func (s *MockStore) DeleteEntity(ctx context.Context, e *types.Entity) error {
	args := s.Called(ctx, e)
	return args.Error(0)
}

// DeleteEntityByName ...
func (s *MockStore) DeleteEntityByName(ctx context.Context, id string) error {
	args := s.Called(ctx, id)
	return args.Error(0)
}

// GetEntities ...
func (s *MockStore) GetEntities(ctx context.Context, pageSize int64, continueToken string) ([]*types.Entity, string, error) {
	args := s.Called(ctx, pageSize, continueToken)
	return args.Get(0).([]*types.Entity), args.String(1), args.Error(2)
}

// GetEntityByName ...
func (s *MockStore) GetEntityByName(ctx context.Context, id string) (*types.Entity, error) {
	args := s.Called(ctx, id)
	return args.Get(0).(*types.Entity), args.Error(1)
}

// UpdateEntity ...
func (s *MockStore) UpdateEntity(ctx context.Context, e *types.Entity) error {
	args := s.Called(ctx, e)
	return args.Error(0)
}

// GetEntityWatcher
func (s *MockStore) GetEntityWatcher(ctx context.Context) <-chan store.WatchEventEntity {
	args := s.Called(ctx)
	return args.Get(0).(<-chan store.WatchEventEntity)
}
