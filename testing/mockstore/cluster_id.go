package mockstore

import (
	"context"
)

// CreateClusterID ...
func (s *MockStore) CreateClusterID(ctx context.Context, id string) error {
	args := s.Called(ctx, id)
	return args.Error(0)
}

// GetClusterID ...
func (s *MockStore) GetClusterID(ctx context.Context) (string, error) {
	args := s.Called(ctx)
	return args.Get(0).(string), args.Error(1)
}
