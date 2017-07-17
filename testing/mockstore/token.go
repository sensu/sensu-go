package mockstore

import "github.com/sensu/sensu-go/types"

// CreateToken ...
func (s *MockStore) CreateToken(claims *types.Claims) error {
	args := s.Called(claims)
	return args.Error(0)
}

// DeleteToken ...
func (s *MockStore) DeleteToken(jti string) error {
	args := s.Called(jti)
	return args.Error(0)
}
