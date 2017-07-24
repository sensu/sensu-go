package mockstore

import "github.com/sensu/sensu-go/types"

// CreateToken ...
func (s *MockStore) CreateToken(claims *types.Claims) error {
	args := s.Called(claims)
	return args.Error(0)
}

// DeleteTokens ...
func (s *MockStore) DeleteTokens(subject string, ids []string) error {
	args := s.Called(subject, ids)
	return args.Error(0)
}

// GetToken ...
func (s *MockStore) GetToken(subject, id string) (*types.Claims, error) {
	args := s.Called(subject, id)
	return args.Get(0).(*types.Claims), args.Error(1)
}
