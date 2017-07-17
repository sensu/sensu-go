package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// GetKeepalive ...
func (s *MockStore) GetKeepalive(ctx context.Context, entityID string) (int64, error) {
	args := s.Called(ctx, entityID)
	return args.Get(0).(int64), args.Error(1)
}

// GetFailingKeepalives ...
func (s *MockStore) GetFailingKeepalives(ctx context.Context) ([]*types.KeepaliveRecord, error) {
	args := s.Called(ctx)
	return args.Get(0).([]*types.KeepaliveRecord), args.Error(1)
}

// UpdateKeepalive ...
func (s *MockStore) UpdateKeepalive(ctx context.Context, entityID string, expiration int64) error {
	args := s.Called(ctx, entityID, expiration)
	return args.Error(0)
}
