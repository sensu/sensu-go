package mockstore

import "github.com/sensu/sensu-go/types"

//// Entities

// GetEntityByID ...
func (s *MockStore) GetEntityByID(org, id string) (*types.Entity, error) {
	args := s.Called(org, id)
	return args.Get(0).(*types.Entity), args.Error(1)
}

// UpdateEntity ...
func (s *MockStore) UpdateEntity(e *types.Entity) error {
	args := s.Called(e)
	return args.Error(0)
}

// DeleteEntity ...
func (s *MockStore) DeleteEntity(e *types.Entity) error {
	args := s.Called(e)
	return args.Error(0)
}

// DeleteEntityByID ...
func (s *MockStore) DeleteEntityByID(org, id string) error {
	args := s.Called(org, id)
	return args.Error(0)
}

// GetEntities ...
func (s *MockStore) GetEntities(org string) ([]*types.Entity, error) {
	args := s.Called(org)
	return args.Get(0).([]*types.Entity), args.Error(1)
}
