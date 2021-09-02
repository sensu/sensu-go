package mockstore

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// GetPipelineByName ...
func (s *MockStore) GetPipelineByName(ctx context.Context, name string) (*corev2.Pipeline, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*corev2.Pipeline), args.Error(1)
}
