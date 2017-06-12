package mockstore

import "github.com/sensu/sensu-go/types"

//// Users

// CreateUser ...
func (s *MockStore) CreateUser(user *types.User) error {
	args := s.Called(user)
	return args.Error(0)
}

// DeleteUserByName ...
func (s *MockStore) DeleteUserByName(username string) error {
	args := s.Called(username)
	return args.Error(0)
}

// GetUser ...
func (s *MockStore) GetUser(username string) (*types.User, error) {
	args := s.Called(username)
	return args.Get(0).(*types.User), args.Error(1)
}

// GetUsers ...
func (s *MockStore) GetUsers() ([]*types.User, error) {
	args := s.Called()
	return args.Get(0).([]*types.User), args.Error(1)
}

// UpdateUser ...
func (s *MockStore) UpdateUser(user *types.User) error {
	args := s.Called(user)
	return args.Error(0)
}
