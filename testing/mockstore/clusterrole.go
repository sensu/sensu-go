package mockstore

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// CreateClusterRole ...
func (s *MockStore) CreateClusterRole(ctx context.Context, clusterRole *corev2.ClusterRole) error {
	args := s.Called(ctx, clusterRole)
	return args.Error(0)
}

// CreateOrUpdateClusterRole ...
func (s *MockStore) CreateOrUpdateClusterRole(ctx context.Context, clusterRole *corev2.ClusterRole) error {
	args := s.Called(ctx, clusterRole)
	return args.Error(0)
}

// DeleteClusterRole ...
func (s *MockStore) DeleteClusterRole(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetClusterRole ...
func (s *MockStore) GetClusterRole(ctx context.Context, name string) (*corev2.ClusterRole, error) {
	args := s.Called(ctx, name)
	err := args.Error(1)

	if clusterRole, ok := args.Get(0).(*corev2.ClusterRole); ok {
		return clusterRole, err
	}
	return nil, err
}

// ListClusterRoles ...
func (s *MockStore) ListClusterRoles(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.ClusterRole, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*corev2.ClusterRole), args.Error(1)
}

// UpdateClusterRole ...
func (s *MockStore) UpdateClusterRole(ctx context.Context, clusterRole *corev2.ClusterRole) error {
	args := s.Called(ctx, clusterRole)
	return args.Error(0)
}
