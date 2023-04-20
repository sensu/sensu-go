package mockstore

import (
	"context"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// AuthenticateUser ...
func (s *MockStore) AuthenticateUser(
	ctx context.Context,
	username, password string,
) (*v2.User, error) {
	args := s.Called(ctx, username, password)
	return args.Get(0).(*v2.User), args.Error(1)
}

// CreateUser ...
func (s *MockStore) CreateUser(ctx context.Context, user *v2.User) error {
	args := s.Called(ctx, user)
	return args.Error(0)
}

// DeleteUser ...
func (s *MockStore) DeleteUser(ctx context.Context, user *v2.User) error {
	args := s.Called(ctx, user)
	return args.Error(0)
}

// GetUser ...
func (s *MockStore) GetUser(ctx context.Context, username string) (*v2.User, error) {
	args := s.Called(ctx, username)
	return args.Get(0).(*v2.User), args.Error(1)
}

// GetUsers ...
func (s *MockStore) GetUsers() ([]*v2.User, error) {
	args := s.Called()
	return args.Get(0).([]*v2.User), args.Error(1)
}

// GetAllUsers ...
func (s *MockStore) GetAllUsers(pred *store.SelectionPredicate) ([]*v2.User, error) {
	args := s.Called(pred)
	return args.Get(0).([]*v2.User), args.Error(1)
}

// UpdateUser ...
func (s *MockStore) UpdateUser(user *v2.User) error {
	args := s.Called(user)
	return args.Error(0)
}
