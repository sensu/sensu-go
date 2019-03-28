package mockstore

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// CreateOrUpdateTessenConfig ...
func (s *MockStore) CreateOrUpdateTessenConfig(ctx context.Context, config *corev2.TessenConfig) error {
	args := s.Called(ctx, config)
	return args.Error(0)
}

// GetTessenConfig ...
func (s *MockStore) GetTessenConfig(ctx context.Context) (*corev2.TessenConfig, error) {
	args := s.Called(ctx)
	return args.Get(0).(*corev2.TessenConfig), args.Error(1)
}
