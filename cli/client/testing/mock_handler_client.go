package testing

import (
	"github.com/sensu/sensu-go/types"
)

func (c *MockClient) ListHandlers() ([]types.Handler, error) {
	args := c.Called()
	return args.Get(0).([]types.Handler), args.Error(1)
}

func (c *MockClient) CreateHandler(h *types.Handler) error {
	args := c.Called(h)
	return args.Error(0)
}

func (c *MockClient) DeleteHandler(h *types.Handler) error {
	args := c.Called(h)
	return args.Error(0)
}
