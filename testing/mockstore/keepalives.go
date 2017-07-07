package mockstore

import "context"

// GetKeepalive ...
func (s *MockStore) GetKeepalive(ctx context.Context, entityID string) (int64, error) {
	args := s.Called(ctx, entityID)
	return args.Get(0).(int64), args.Error(1)
}

// UpdateKeepalive ...
func (s *MockStore) UpdateKeepalive(ctx context.Context, entityID string, expiration int64) error {
	args := s.Called(ctx, entityID, expiration)
	return args.Error(0)
}
