package mockstore

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// CreateClusterRoleBinding ...
func (s *MockStore) CreateClusterRoleBinding(ctx context.Context, ClusterRoleBinding *corev2.ClusterRoleBinding) error {
	args := s.Called(ctx, ClusterRoleBinding)
	return args.Error(0)
}

// CreateOrUpdateClusterRoleBinding ...
func (s *MockStore) CreateOrUpdateClusterRoleBinding(ctx context.Context, ClusterRoleBinding *corev2.ClusterRoleBinding) error {
	args := s.Called(ctx, ClusterRoleBinding)
	return args.Error(0)
}

// DeleteClusterRoleBinding ...
func (s *MockStore) DeleteClusterRoleBinding(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetClusterRoleBinding ...
func (s *MockStore) GetClusterRoleBinding(ctx context.Context, name string) (*corev2.ClusterRoleBinding, error) {
	args := s.Called(ctx, name)
	err := args.Error(1)

	if clusterRoleBinding, ok := args.Get(0).(*corev2.ClusterRoleBinding); ok {
		return clusterRoleBinding, err
	}
	return nil, err
}

// ListClusterRoleBindings ...
func (s *MockStore) ListClusterRoleBindings(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.ClusterRoleBinding, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*corev2.ClusterRoleBinding), args.Error(1)
}

// UpdateClusterRoleBinding ...
func (s *MockStore) UpdateClusterRoleBinding(ctx context.Context, clusterRoleBinding *corev2.ClusterRoleBinding) error {
	args := s.Called(ctx, clusterRoleBinding)
	return args.Error(0)
}
