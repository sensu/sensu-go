package testing

import "github.com/sensu/sensu-go/types"

// Health ...
func (c *MockClient) Health() (*types.HealthResponse, error) {
	args := c.Called()
	return args.Get(0).(*types.HealthResponse), args.Error(1)
}
