package testing

import "github.com/sensu/sensu-go/types"

// CreateMutator for use with mock package
func (c *MockClient) CreateMutator(m *types.Mutator) error {
	args := c.Called(m)
	return args.Error(0)
}

// DeleteMutator for use with mock package
func (c *MockClient) DeleteMutator(m *types.Mutator) error {
	args := c.Called(m)
	return args.Error(0)
}

// FetchMutator for use with mock package
func (c *MockClient) FetchMutator(name string) (*types.Mutator, error) {
	args := c.Called(name)
	return args.Get(0).(*types.Mutator), args.Error(1)
}

// UpdateMutator for use with mock package
func (c *MockClient) UpdateMutator(m *types.Mutator) error {
	args := c.Called(m)
	return args.Error(0)
}

// ListMutators for use with mock lib
func (c *MockClient) ListMutators(org string) ([]types.Mutator, error) {
	args := c.Called(org)
	return args.Get(0).([]types.Mutator), args.Error(1)
}
