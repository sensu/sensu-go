package mockstore

import "github.com/sensu/sensu-go/types"

//// Handlers

// GetHandlers ...
func (s *MockStore) GetHandlers(org string) ([]*types.Handler, error) {
	args := s.Called(org)
	return args.Get(0).([]*types.Handler), args.Error(1)
}

// GetHandlerByName ...
func (s *MockStore) GetHandlerByName(org, name string) (*types.Handler, error) {
	args := s.Called(org, name)
	return args.Get(0).(*types.Handler), args.Error(1)
}

// DeleteHandlerByName ...
func (s *MockStore) DeleteHandlerByName(org, name string) error {
	args := s.Called(org, name)
	return args.Error(0)
}

// UpdateHandler ...
func (s *MockStore) UpdateHandler(handler *types.Handler) error {
	args := s.Called(handler)
	return args.Error(0)
}
