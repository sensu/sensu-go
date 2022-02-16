package mockstore

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
)

// CreateResource ...
func (s *MockStore) CreateResource(ctx context.Context, resource corev2.Resource) error {
	args := s.Called(ctx, resource)
	return args.Error(0)
}

// CreateOrUpdateResource ...
func (s *MockStore) CreateOrUpdateResource(ctx context.Context, resource corev2.Resource, prev ...corev2.Resource) error {
	args := s.Called(ctx, resource, prev)
	return args.Error(0)
}

// DeleteResource ...
func (s *MockStore) DeleteResource(ctx context.Context, kind, name string) error {
	args := s.Called(ctx, kind, name)
	return args.Error(0)
}

// GetResource ...
func (s *MockStore) GetResource(ctx context.Context, name string, resource corev2.Resource) error {
	args := s.Called(ctx, name, resource)
	return args.Error(0)
}

// ListResources ...
func (s *MockStore) ListResources(ctx context.Context, kind string, list interface{}, pred *store.SelectionPredicate) error {
	args := s.Called(ctx, kind, list, pred)
	return args.Error(0)
}

// PatchResource ...
func (s *MockStore) PatchResource(ctx context.Context, resource corev2.Resource, name string, patcher patch.Patcher, condition *store.ETagCondition) error {
	args := s.Called(ctx, resource, name, patcher, condition)
	return args.Error(0)
}
