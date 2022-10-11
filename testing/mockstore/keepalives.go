package mockstore

import (
	"context"

	corev2 "github.com/sensu/core/v2"
)

// DeleteFailingKeepalive ...
func (s *MockStore) DeleteFailingKeepalive(ctx context.Context, entity *corev2.Entity) error {
	args := s.Called(ctx, entity)
	return args.Error(0)
}

// GetFailingKeepalives ...
func (s *MockStore) GetFailingKeepalives(ctx context.Context) ([]*corev2.KeepaliveRecord, error) {
	args := s.Called(ctx)
	return args.Get(0).([]*corev2.KeepaliveRecord), args.Error(1)
}

// UpdateFailingKeepalive ...
func (s *MockStore) UpdateFailingKeepalive(ctx context.Context, entity *corev2.Entity, expiration int64) error {
	args := s.Called(ctx, entity, expiration)
	return args.Error(0)
}
