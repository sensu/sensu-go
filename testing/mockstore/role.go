package mockstore

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// CreateRole ...
func (s *MockStore) CreateRole(ctx context.Context, role *corev2.Role) error {
	args := s.Called(ctx, role)
	return args.Error(0)
}

// CreateOrUpdateRole ...
func (s *MockStore) CreateOrUpdateRole(ctx context.Context, role *corev2.Role) error {
	args := s.Called(ctx, role)
	return args.Error(0)
}

// DeleteRole ...
func (s *MockStore) DeleteRole(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetRole ...
func (s *MockStore) GetRole(ctx context.Context, name string) (*corev2.Role, error) {
	args := s.Called(ctx, name)
	err := args.Error(1)

	if role, ok := args.Get(0).(*corev2.Role); ok {
		return role, err
	}
	return nil, err
}

// ListRoles ...
func (s *MockStore) ListRoles(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Role, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*corev2.Role), args.Error(1)
}

// UpdateRole ...
func (s *MockStore) UpdateRole(ctx context.Context, role *corev2.Role) error {
	args := s.Called(ctx, role)
	return args.Error(0)
}
