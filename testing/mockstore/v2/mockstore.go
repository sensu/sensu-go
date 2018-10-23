package mockstore

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockStore is a store used for testing. When using the MockStore in unit
// tests, stub out the behavior you wish to test against by assigning the
// appropriate function to the appropriate Func field. If you have forgotten
// to stub a particular function, the program will panic.
type MockStore struct {
	mock.Mock
}

// Create ...
func (s *MockStore) Create(ctx context.Context, key string, objPtr interface{}) error {
	args := s.Called(ctx, key, objPtr)
	return args.Error(0)
}

// CreateOrUpdate ...
func (s *MockStore) CreateOrUpdate(ctx context.Context, key string, objPtr interface{}) error {
	args := s.Called(ctx, key, objPtr)
	return args.Error(0)
}

// Get ...
func (s *MockStore) Get(ctx context.Context, key string, objPtr interface{}) error {
	args := s.Called(ctx, key, objPtr)
	return args.Error(0)
}

// List ...
func (s *MockStore) List(ctx context.Context, key string, objsPtr interface{}) error {
	args := s.Called(ctx, key, objsPtr)
	return args.Error(0)
}

// Update ...
func (s *MockStore) Update(ctx context.Context, key string, objPtr interface{}) error {
	args := s.Called(ctx, key, objPtr)
	return args.Error(0)
}
