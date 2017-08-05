package testing

import "github.com/sensu/sensu-go/types"

// CreateEnvironment for use with mock lib
func (c *MockClient) CreateEnvironment(org string, env *types.Environment) error {
	args := c.Called(org, env)
	return args.Error(0)
}

// DeleteEnvironment for use with mock lib
func (c *MockClient) DeleteEnvironment(org, env string) error {
	args := c.Called(org, env)
	return args.Error(0)
}

// ListEnvironments for use with mock lib
func (c *MockClient) ListEnvironments(org string) ([]types.Environment, error) {
	args := c.Called(org)
	return args.Get(0).([]types.Environment), args.Error(1)
}
