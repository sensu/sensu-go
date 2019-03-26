package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// CreateNamespace ...
func (s *MockStore) CreateNamespace(ctx context.Context, org *types.Namespace) error {
	args := s.Called(ctx, org)
	return args.Error(0)
}

// DeleteNamespace ...
func (s *MockStore) DeleteNamespace(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// ListNamespaces ...
func (s *MockStore) ListNamespaces(ctx context.Context, pageSize int64, continueToken string) ([]*types.Namespace, string, error) {
	args := s.Called(ctx, pageSize, continueToken)
	return args.Get(0).([]*types.Namespace), args.String(1), args.Error(2)
}

// GetNamespace ...
func (s *MockStore) GetNamespace(ctx context.Context, name string) (*types.Namespace, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*types.Namespace), args.Error(1)
}

// UpdateNamespace ...
func (s *MockStore) UpdateNamespace(ctx context.Context, org *types.Namespace) error {
	args := s.Called(ctx, org)
	return args.Error(0)
}
