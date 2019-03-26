package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// CreateRole ...
func (s *MockStore) CreateRole(ctx context.Context, role *types.Role) error {
	args := s.Called(ctx, role)
	return args.Error(0)
}

// CreateOrUpdateRole ...
func (s *MockStore) CreateOrUpdateRole(ctx context.Context, role *types.Role) error {
	args := s.Called(ctx, role)
	return args.Error(0)
}

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
func (s *MockStore) ListRoles(ctx context.Context, pageSize int64, continueToken string) ([]*types.Role, string, error) {
	args := s.Called(ctx, pageSize, continueToken)
	return args.Get(0).([]*types.Role), args.String(1), args.Error(2)
}

// UpdateRole ...
func (s *MockStore) UpdateRole(ctx context.Context, role *types.Role) error {
	args := s.Called(ctx, role)
	return args.Error(0)
}
