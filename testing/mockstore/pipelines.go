package mockstore

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// DeletePipelineByName ...
func (s *MockStore) DeletePipelineByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetPipelines ...
func (s *MockStore) GetPipelines(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Pipeline, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*corev2.Pipeline), args.Error(1)
}

// GetPipelineByName ...
func (s *MockStore) GetPipelineByName(ctx context.Context, name string) (*corev2.Pipeline, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*corev2.Pipeline), args.Error(1)
}

// UpdatePipeline ...
func (s *MockStore) UpdatePipeline(ctx context.Context, pipeline *corev2.Pipeline) error {
	args := s.Called(ctx, pipeline)
	return args.Error(0)
}
