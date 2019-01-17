package mockstore

import (
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

// CreateToken ...
func (s *MockStore) AllowTokens(tokens ...*jwt.Token) error {
	args := s.Called(tokens)
	return args.Error(0)
}

// RevokeTokens ...
func (s *MockStore) RevokeTokens(claims ...*v2.Claims) error {
	args := s.Called(claims)
	return args.Error(0)
}

// GetToken ...
func (s *MockStore) GetToken(subject, id string) (*types.Claims, error) {
	args := s.Called(subject, id)
	return args.Get(0).(*types.Claims), args.Error(1)
}
