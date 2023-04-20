package mockstore

import (
	"context"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// DeleteEntity ...
func (s *MockStore) DeleteEntity(ctx context.Context, e *v2.Entity) error {
	args := s.Called(ctx, e)
	return args.Error(0)
}

// DeleteEntityByName ...
func (s *MockStore) DeleteEntityByName(ctx context.Context, id string) error {
	args := s.Called(ctx, id)
	return args.Error(0)
}

// GetEntities ...
func (s *MockStore) GetEntities(ctx context.Context, pred *store.SelectionPredicate) ([]*v2.Entity, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*v2.Entity), args.Error(1)
}

// GetEntityByName ...
func (s *MockStore) GetEntityByName(ctx context.Context, id string) (*v2.Entity, error) {
	args := s.Called(ctx, id)
	return args.Get(0).(*v2.Entity), args.Error(1)
}

// UpdateEntity ...
func (s *MockStore) UpdateEntity(ctx context.Context, e *v2.Entity) error {
	args := s.Called(ctx, e)
	return args.Error(0)
}
