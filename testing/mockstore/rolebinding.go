package mockstore

import (
	"context"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// CreateRoleBinding ...
func (s *MockStore) CreateRoleBinding(ctx context.Context, RoleBinding *v2.RoleBinding) error {
	args := s.Called(ctx, RoleBinding)
	return args.Error(0)
}

// CreateOrUpdateRoleBinding ...
func (s *MockStore) CreateOrUpdateRoleBinding(ctx context.Context, RoleBinding *v2.RoleBinding) error {
	args := s.Called(ctx, RoleBinding)
	return args.Error(0)
}

// DeleteRoleBinding ...
func (s *MockStore) DeleteRoleBinding(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetRoleBinding ...
func (s *MockStore) GetRoleBinding(ctx context.Context, name string) (*v2.RoleBinding, error) {
	args := s.Called(ctx, name)
	err := args.Error(1)

	if roleBinding, ok := args.Get(0).(*v2.RoleBinding); ok {
		return roleBinding, err
	}
	return nil, err
}

// ListRoleBindings ...
func (s *MockStore) ListRoleBindings(ctx context.Context, pred *store.SelectionPredicate) ([]*v2.RoleBinding, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*v2.RoleBinding), args.Error(1)
}

// UpdateRoleBinding ...
func (s *MockStore) UpdateRoleBinding(ctx context.Context, roleBinding *v2.RoleBinding) error {
	args := s.Called(ctx, roleBinding)
	return args.Error(0)
}
