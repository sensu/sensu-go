package testing

import "github.com/sensu/sensu-go/types"

// PutGeneric ...
func (c *MockClient) PutResource(r types.Resource) error {
	args := c.Called(r)
	return args.Error(0)
}
