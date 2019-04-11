package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// CreateRoleBinding ...
func (s *MockStore) CreateRoleBinding(ctx context.Context, RoleBinding *types.RoleBinding) error {
	args := s.Called(ctx, RoleBinding)
	return args.Error(0)
}

// CreateOrUpdateRoleBinding ...
func (s *MockStore) CreateOrUpdateRoleBinding(ctx context.Context, RoleBinding *types.RoleBinding) error {
	args := s.Called(ctx, RoleBinding)
	return args.Error(0)
}

// DeleteRoleBinding ...
func (s *MockStore) DeleteRoleBinding(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetRoleBinding ...
func (s *MockStore) GetRoleBinding(ctx context.Context, name string) (*types.RoleBinding, error) {
	args := s.Called(ctx, name)
	err := args.Error(1)

	if roleBinding, ok := args.Get(0).(*types.RoleBinding); ok {
		return roleBinding, err
	}
	return nil, err
}

// ListRoleBindings ...
func (s *MockStore) ListRoleBindings(ctx context.Context, pred *store.SelectionPredicate) ([]*types.RoleBinding, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*types.RoleBinding), args.Error(1)
}

// UpdateRoleBinding ...
func (s *MockStore) UpdateRoleBinding(ctx context.Context, roleBinding *types.RoleBinding) error {
	args := s.Called(ctx, roleBinding)
	return args.Error(0)
}
