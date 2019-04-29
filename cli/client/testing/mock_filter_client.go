package testing

import (
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// CreateFilter for use with mock lib
func (c *MockClient) CreateFilter(filter *types.EventFilter) error {
	args := c.Called(filter)
	return args.Error(0)
}

// DeleteFilter for use with mock lib
func (c *MockClient) DeleteFilter(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}

// FetchFilter for use with mock lib
func (c *MockClient) FetchFilter(name string) (*types.EventFilter, error) {
	args := c.Called(name)
	return args.Get(0).(*types.EventFilter), args.Error(1)
}

// ListFilters for use with mock lib
func (c *MockClient) ListFilters(namespace string, options client.ListOptions) ([]corev2.EventFilter, string, error) {
	args := c.Called(namespace, options)
	return args.Get(0).([]corev2.EventFilter), args.Get(1).(string), args.Error(2)
}

// UpdateFilter for use with mock lib
func (c *MockClient) UpdateFilter(filter *types.EventFilter) error {
	args := c.Called(filter)
	return args.Error(0)
}
