package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// DeleteFailingKeepalive ...
func (s *MockStore) DeleteFailingKeepalive(ctx context.Context, entity *types.Entity) error {
	args := s.Called(ctx, entity)
	return args.Error(0)
}

// GetFailingKeepalives ...
func (s *MockStore) GetFailingKeepalives(ctx context.Context) ([]*types.KeepaliveRecord, error) {
	args := s.Called(ctx)
	return args.Get(0).([]*types.KeepaliveRecord), args.Error(1)
}

// UpdateFailingKeepalive ...
func (s *MockStore) UpdateFailingKeepalive(ctx context.Context, entity *types.Entity, expiration int64) error {
	args := s.Called(ctx, entity, expiration)
	return args.Error(0)
}
