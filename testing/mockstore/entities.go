package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// DeleteEntity ...
func (s *MockStore) DeleteEntity(e *types.Entity) error {
	args := s.Called(e)
	return args.Error(0)
}

// DeleteEntityByID ...
func (s *MockStore) DeleteEntityByID(ctx context.Context, id string) error {
	args := s.Called(ctx, id)
	return args.Error(0)
}

// GetEntities ...
func (s *MockStore) GetEntities(ctx context.Context) ([]*types.Entity, error) {
	args := s.Called(ctx)
	return args.Get(0).([]*types.Entity), args.Error(1)
}

// GetEntityByID ...
func (s *MockStore) GetEntityByID(ctx context.Context, id string) (*types.Entity, error) {
	args := s.Called(ctx, id)
	return args.Get(0).(*types.Entity), args.Error(1)
}

// UpdateEntity ...
func (s *MockStore) UpdateEntity(e *types.Entity) error {
	args := s.Called(e)
	return args.Error(0)
}
