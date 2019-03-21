package testing

import (
	"github.com/sensu/sensu-go/types"
)

// ListHandlers for use with mock package
func (c *MockClient) ListHandlers(namespace string) ([]types.Handler, error) {
	args := c.Called(namespace)
	return args.Get(0).([]types.Handler), args.Error(1)
}

// CreateHandler for use with mock package
func (c *MockClient) CreateHandler(h *types.Handler) error {
	args := c.Called(h)
	return args.Error(0)
}

// DeleteHandler for use with mock package
func (c *MockClient) DeleteHandler(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}

// FetchHandler for use with mock lib
func (c *MockClient) FetchHandler(name string) (*types.Handler, error) {
	args := c.Called(name)
	return args.Get(0).(*types.Handler), args.Error(1)
}

// UpdateHandler for use with mock lib
func (c *MockClient) UpdateHandler(h *types.Handler) error {
	args := c.Called(h)
	return args.Error(0)
}
