package mockstore

import (
	"context"

	corev2 "github.com/sensu/core/v2"
)

// GetFallbackPipelines retrieves fallback pipelines by name.
func (s *MockStore) GetFallbackPipelines(ctx context.Context, name string) ([]*corev2.Pipeline, error) {
	// TODO - implement the logic to retrieve fallback pipelines from the mock store
	args := s.Called(ctx, name)
	if pipelines, ok := args.Get(0).([]*corev2.Pipeline); ok {
		return pipelines, args.Error(1)
	}
	return nil, args.Error(1)
}
