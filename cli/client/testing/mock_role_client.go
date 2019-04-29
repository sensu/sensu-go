package testing

import (
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// CreateRole ...
func (c *MockClient) CreateRole(check *types.Role) error {
	args := c.Called(check)
	return args.Error(0)
}

// FetchRole ...
func (c *MockClient) FetchRole(name string) (*types.Role, error) {
	args := c.Called(name)
	return args.Get(0).(*types.Role), args.Error(1)
}

// DeleteRole ...
func (c *MockClient) DeleteRole(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}

// ListRoles ...
func (c *MockClient) ListRoles(namespace string, options client.ListOptions) ([]corev2.Role, string, error) {
	args := c.Called(namespace, options)
	return args.Get(0).([]corev2.Role), args.Get(1).(string), args.Error(2)
}
