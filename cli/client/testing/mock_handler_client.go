package testing

import (
	"github.com/sensu/sensu-go/types"
)

// ListHandlers for use with mock package
func (c *MockClient) ListHandlers() ([]types.Handler, error) {
	args := c.Called()
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
