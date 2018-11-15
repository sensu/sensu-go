package mockstore

import (
	"context"

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
func (s *MockStore) GetEntities(ctx context.Context) ([]*types.Entity, error) {
	args := s.Called(ctx)
	return args.Get(0).([]*types.Entity), args.Error(1)
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
