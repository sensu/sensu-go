package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// CreateClusterRole ...
func (s *MockStore) CreateClusterRole(ctx context.Context, clusterRole *types.ClusterRole) error {
	args := s.Called(ctx, clusterRole)
	return args.Error(0)
}

// CreateOrUpdateClusterRole ...
func (s *MockStore) CreateOrUpdateClusterRole(ctx context.Context, clusterRole *types.ClusterRole) error {
	args := s.Called(ctx, clusterRole)
	return args.Error(0)
}

// DeleteClusterRole ...
func (s *MockStore) DeleteClusterRole(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetClusterRole ...
func (s *MockStore) GetClusterRole(ctx context.Context, name string) (*types.ClusterRole, error) {
	args := s.Called(ctx, name)
	err := args.Error(1)

	if clusterRole, ok := args.Get(0).(*types.ClusterRole); ok {
		return clusterRole, err
	}
	return nil, err
}

// ListClusterRoles ...
func (s *MockStore) ListClusterRoles(ctx context.Context, pageSize int64, continueToken string) ([]*types.ClusterRole, string, error) {
	args := s.Called(ctx, pageSize, continueToken)
	return args.Get(0).([]*types.ClusterRole), args.String(1), args.Error(2)
}

// UpdateClusterRole ...
func (s *MockStore) UpdateClusterRole(ctx context.Context, clusterRole *types.ClusterRole) error {
	args := s.Called(ctx, clusterRole)
	return args.Error(0)
}
