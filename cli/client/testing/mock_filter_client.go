package testing

import (
	corev2 "github.com/sensu/core/v2"
)

// CreateFilter for use with mock lib
func (c *MockClient) CreateFilter(filter *corev2.EventFilter) error {
	args := c.Called(filter)
	return args.Error(0)
}

// DeleteFilter for use with mock lib
func (c *MockClient) DeleteFilter(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}

// FetchFilter for use with mock lib
func (c *MockClient) FetchFilter(name string) (*corev2.EventFilter, error) {
	args := c.Called(name)
	return args.Get(0).(*corev2.EventFilter), args.Error(1)
}

// UpdateFilter for use with mock lib
func (c *MockClient) UpdateFilter(filter *corev2.EventFilter) error {
	args := c.Called(filter)
	return args.Error(0)
}
