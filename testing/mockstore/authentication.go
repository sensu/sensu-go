package mockstore

//// Authentication

// CreateJWTSecret ...
func (s *MockStore) CreateJWTSecret(secret []byte) error {
	args := s.Called()
	return args.Error(0)
}

// GetJWTSecret ...
func (s *MockStore) GetJWTSecret() ([]byte, error) {
	args := s.Called()
	return []byte(args.String(0)), args.Error(1)
}

// UpdateJWTSecret ...
func (s *MockStore) UpdateJWTSecret(secret []byte) error {
	args := s.Called(secret)
	return args.Error(0)
}
