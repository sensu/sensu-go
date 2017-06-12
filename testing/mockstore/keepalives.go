package mockstore

//// Keepalives

// UpdateKeepalive ...
func (s *MockStore) UpdateKeepalive(org, entityID string, expiration int64) error {
	args := s.Called(org, entityID, expiration)
	return args.Error(0)
}

// GetKeepalive ...
func (s *MockStore) GetKeepalive(org, entityID string) (int64, error) {
	args := s.Called(org, entityID)
	return args.Get(0).(int64), args.Error(1)
}
