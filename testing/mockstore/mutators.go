package mockstore

import (
	"context"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// DeleteMutatorByName ...
func (s *MockStore) DeleteMutatorByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetMutators ...
func (s *MockStore) GetMutators(ctx context.Context, pred *store.SelectionPredicate) ([]*v2.Mutator, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*v2.Mutator), args.Error(1)
}

// GetMutatorByName ...
func (s *MockStore) GetMutatorByName(ctx context.Context, name string) (*v2.Mutator, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*v2.Mutator), args.Error(1)
}

// UpdateMutator ...
func (s *MockStore) UpdateMutator(ctx context.Context, mutator *v2.Mutator) error {
	args := s.Called(mutator)
	return args.Error(0)
}
