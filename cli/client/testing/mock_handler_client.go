package testing

import (
	"github.com/sensu/sensu-go/types"
)

// ListHandlers for use with mock package
func (c *MockClient) ListHandlers(org string) ([]types.Handler, error) {
	args := c.Called(org)
	return args.Get(0).([]types.Handler), args.Error(1)
}

// CreateHandler for use with mock package
func (c *MockClient) CreateHandler(h *types.Handler) error {
	args := c.Called(h)
	return args.Error(0)
}

// DeleteHandler for use with mock package
func (c *MockClient) DeleteHandler(h *types.Handler) error {
	args := c.Called(h)
	return args.Error(0)
}

// FetchHandler for use with mock lib
func (c *MockClient) FetchHandler(name string) (*types.Handler, error) {
	args := c.Called(name)
	return args.Get(0).(*types.Handler), args.Error(1)
}
