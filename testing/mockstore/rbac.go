package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// GetRoles ...
func (s *MockStore) GetRoles(ctx context.Context) ([]*types.Role, error) {
	args := s.Called(ctx)
	return args.Get(0).([]*types.Role), args.Error(1)
}

// GetRoleByName ...
func (s *MockStore) GetRoleByName(ctx context.Context, name string) (*types.Role, error) {
	args := s.Called(ctx, name)
	err := args.Error(1)

	if role, ok := args.Get(0).(*types.Role); ok {
		return role, err
	}
	return nil, err
}

// UpdateRole ...
func (s *MockStore) UpdateRole(ctx context.Context, role *types.Role) error {
	args := s.Called(ctx, role)
	return args.Error(0)
}

// DeleteRoleByName ...
func (s *MockStore) DeleteRoleByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}
