package mockstore

import (
	"context"

	v2 "github.com/sensu/core/v2"
)

// DeleteFailingKeepalive ...
func (s *MockStore) DeleteFailingKeepalive(ctx context.Context, entity *v2.Entity) error {
	args := s.Called(ctx, entity)
	return args.Error(0)
}

// GetFailingKeepalives ...
func (s *MockStore) GetFailingKeepalives(ctx context.Context) ([]*v2.KeepaliveRecord, error) {
	args := s.Called(ctx)
	return args.Get(0).([]*v2.KeepaliveRecord), args.Error(1)
}

// UpdateFailingKeepalive ...
func (s *MockStore) UpdateFailingKeepalive(ctx context.Context, entity *v2.Entity, expiration int64) error {
	args := s.Called(ctx, entity, expiration)
	return args.Error(0)
}
