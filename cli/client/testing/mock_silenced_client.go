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
func (c *MockClient) DeleteSilenced(subscription, check string) error {
	args := c.Called(subscription, check)
	return args.Error(0)
}

// FetchSilenced for use with mock lib
func (c *MockClient) FetchSilenced(subscription, check string) (*types.Silenced, error) {
	args := c.Called(subscription, check)
	return args.Get(0).(*types.Silenced), args.Error(1)
}

// ListSilenceds for use with mock lib
func (c *MockClient) ListSilenceds(org string) ([]types.Silenced, error) {
	args := c.Called(org)
	return args.Get(0).([]types.Silenced), args.Error(1)
}
