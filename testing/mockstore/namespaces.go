package mockstore

import (
	"context"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// CreateNamespace ...
func (s *MockStore) CreateNamespace(ctx context.Context, org *v2.Namespace) error {
	args := s.Called(ctx, org)
	return args.Error(0)
}

// DeleteNamespace ...
func (s *MockStore) DeleteNamespace(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// ListNamespaces ...
func (s *MockStore) ListNamespaces(ctx context.Context, pred *store.SelectionPredicate) ([]*v2.Namespace, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*v2.Namespace), args.Error(1)
}

// GetNamespace ...
func (s *MockStore) GetNamespace(ctx context.Context, name string) (*v2.Namespace, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*v2.Namespace), args.Error(1)
}

// UpdateNamespace ...
func (s *MockStore) UpdateNamespace(ctx context.Context, org *v2.Namespace) error {
	args := s.Called(ctx, org)
	return args.Error(0)
}
