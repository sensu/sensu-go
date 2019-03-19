package testing

import "github.com/sensu/sensu-go/types"

// CreateSilenced for use with mock lib
func (c *MockClient) CreateSilenced(silenced *types.Silenced) error {
	args := c.Called(silenced)
	return args.Error(0)
}

// UpdateSilenced for use with mock lib
func (c *MockClient) UpdateSilenced(silenced *types.Silenced) error {
	args := c.Called(silenced)
	return args.Error(0)
}

// DeleteSilenced for use with mock lib
func (c *MockClient) DeleteSilenced(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}

// FetchSilenced for use with mock lib
func (c *MockClient) FetchSilenced(id string) (*types.Silenced, error) {
	args := c.Called(id)
	return args.Get(0).(*types.Silenced), args.Error(1)
}

// ListSilenceds for use with mock lib
func (c *MockClient) ListSilenceds(namespace, sub, check string) ([]types.Silenced, error) {
	args := c.Called(namespace, sub, check)
	return args.Get(0).([]types.Silenced), args.Error(1)
}
