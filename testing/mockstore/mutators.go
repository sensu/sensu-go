package mockstore

import "github.com/sensu/sensu-go/types"

//// Mutators

// GetMutators ...
func (s *MockStore) GetMutators() ([]*types.Mutator, error) {
	args := s.Called()
	return args.Get(0).([]*types.Mutator), args.Error(1)
}

// GetMutatorByName ...
func (s *MockStore) GetMutatorByName(name string) (*types.Mutator, error) {
	args := s.Called(name)
	return args.Get(0).(*types.Mutator), args.Error(1)
}

// DeleteMutatorByName ...
func (s *MockStore) DeleteMutatorByName(name string) error {
	args := s.Called(name)
	return args.Error(0)
}

// UpdateMutator ...
func (s *MockStore) UpdateMutator(mutator *types.Mutator) error {
	args := s.Called(mutator)
	return args.Error(0)
}
