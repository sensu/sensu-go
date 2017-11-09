package testing

import "github.com/sensu/sensu-go/types"

// CreateFilter for use with mock lib
func (c *MockClient) CreateFilter(filter *types.EventFilter) error {
	args := c.Called(filter)
	return args.Error(0)
}

// DeleteFilter for use with mock lib
func (c *MockClient) DeleteFilter(filter *types.EventFilter) error {
	args := c.Called(filter)
	return args.Error(0)
}

// FetchFilter for use with mock lib
func (c *MockClient) FetchFilter(name string) (*types.EventFilter, error) {
	args := c.Called(name)
	return args.Get(0).(*types.EventFilter), args.Error(1)
}

// ListFilters for use with mock lib
func (c *MockClient) ListFilters(org string) ([]types.EventFilter, error) {
	args := c.Called(org)
	return args.Get(0).([]types.EventFilter), args.Error(1)
}
