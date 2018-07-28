package testing

import "github.com/sensu/sensu-go/types"

func (c *MockClient) Health() ([]*types.ClusterHealth, error) {
	args := c.Called()
	return args.Get(0).([]*types.ClusterHealth), args.Error(1)
}
