package mockstore

//// Keepalives

// UpdateKeepalive ...
func (s *MockStore) UpdateKeepalive(entityID string, expiration int64) error {
	args := s.Called(entityID, expiration)
	return args.Error(0)
}

// GetKeepalive ...
func (s *MockStore) GetKeepalive(entityID string) (int64, error) {
	args := s.Called(entityID)
	return args.Get(0).(int64), args.Error(1)
}
