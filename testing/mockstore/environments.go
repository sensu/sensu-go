package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// DeleteEnvironment ...
func (s *MockStore) DeleteEnvironment(ctx context.Context, org, env string) error {
	args := s.Called(ctx, org, env)
	return args.Error(0)
}

// GetEnvironment ...
func (s *MockStore) GetEnvironment(ctx context.Context, org, env string) (*types.Environment, error) {
	args := s.Called(ctx, org, env)
	return args.Get(0).(*types.Environment), args.Error(1)
}

// GetEnvironments ...
func (s *MockStore) GetEnvironments(ctx context.Context, org string) ([]*types.Environment, error) {
	args := s.Called(ctx, org)
	return args.Get(0).([]*types.Environment), args.Error(1)
}

// UpdateEnvironment ...
func (s *MockStore) UpdateEnvironment(ctx context.Context, org string, env *types.Environment) error {
	args := s.Called(ctx, org, env)
	return args.Error(0)
}
