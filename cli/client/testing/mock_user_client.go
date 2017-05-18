package testing

import "github.com/sensu/sensu-go/types"

// CreateUser for use with mock lib
func (c *MockClient) CreateUser(check *types.User) error {
	args := c.Called(check)
	return args.Error(0)
}
