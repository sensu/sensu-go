package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/backend/tessen"
)

// CreateOrUpdateTessenConfig ...
func (s *MockStore) CreateOrUpdateTessenConfig(ctx context.Context, config *tessen.Config) error {
	args := s.Called(ctx, config)
	return args.Error(0)
}

// GetTessenConfig ...
func (s *MockStore) GetTessenConfig(ctx context.Context) (*tessen.Config, error) {
	args := s.Called(ctx)
	return args.Get(0).(*tessen.Config), args.Error(1)
}
