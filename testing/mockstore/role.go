package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// DeleteRole ...
func (s *MockStore) DeleteRole(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetRole ...
func (s *MockStore) GetRole(ctx context.Context, name string) (*types.Role, error) {
	args := s.Called(ctx, name)
	err := args.Error(1)

	if role, ok := args.Get(0).(*types.Role); ok {
		return role, err
	}
	return nil, err
}

// ListRoles ...
func (s *MockStore) ListRoles(ctx context.Context) ([]*types.Role, error) {
	args := s.Called(ctx)
	return args.Get(0).([]*types.Role), args.Error(1)
}

// UpdateRole ...
func (s *MockStore) UpdateRole(ctx context.Context, role *types.Role) error {
	args := s.Called(ctx, role)
	return args.Error(0)
}
