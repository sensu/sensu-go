package testing

import (
	v2 "github.com/sensu/core/v2"
	// Health ...
)

func (c *MockClient) Health() (*v2.HealthResponse, error) {
	args := c.Called()
	return args.Get(0).(*v2.HealthResponse), args.Error(1)
}
