package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// GetClusterHealth...
func (s *MockStore) GetClusterHealth(ctx context.Context) *types.HealthResponse {
	args := s.Called(ctx)
	return args.Get(0).(*types.HealthResponse)
}
