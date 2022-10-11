package testing

import corev2 "github.com/sensu/core/v2"

// Health ...
func (c *MockClient) Health() (*corev2.HealthResponse, error) {
	args := c.Called()
	return args.Get(0).(*corev2.HealthResponse), args.Error(1)
}
