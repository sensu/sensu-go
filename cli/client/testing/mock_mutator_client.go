package testing

import "github.com/sensu/sensu-go/types"

// ListMutators for use with mock lib
func (c *MockClient) ListMutators() ([]types.Mutator, error) {
	args := c.Called()
	return args.Get(0).([]types.Mutator), args.Error(1)
}

// CreateMutator for use with mock package
func (c *MockClient) CreateMutator(m *types.Mutator) error {
	args := c.Called(m)
	return args.Error(0)
}
