package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// AuthenticateUser ...
func (s *MockStore) AuthenticateUser(
	ctx context.Context,
	username, password string,
) (*types.User, error) {
	args := s.Called(ctx, username, password)
	return args.Get(0).(*types.User), args.Error(1)
}

// CreateUser ...
func (s *MockStore) CreateUser(user *types.User) error {
	args := s.Called(user)
	return args.Error(0)
}

// DeleteUser ...
func (s *MockStore) DeleteUser(ctx context.Context, user *types.User) error {
	args := s.Called(ctx, user)
	return args.Error(0)
}

// GetUser ...
func (s *MockStore) GetUser(ctx context.Context, username string) (*types.User, error) {
	args := s.Called(ctx, username)
	return args.Get(0).(*types.User), args.Error(1)
}

// GetUsers ...
func (s *MockStore) GetUsers() ([]*types.User, error) {
	args := s.Called()
	return args.Get(0).([]*types.User), args.Error(1)
}

// GetAllUsers ...
func (s *MockStore) GetAllUsers(pred *store.SelectionPredicate) ([]*types.User, error) {
	args := s.Called(pred)
	return args.Get(0).([]*types.User), args.Error(1)
}

// UpdateUser ...
func (s *MockStore) UpdateUser(user *types.User) error {
	args := s.Called(user)
	return args.Error(0)
}
