package testing

import (
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// ListExtensions ...
func (c *MockClient) ListExtensions(namespace string, options client.ListOptions) ([]corev2.Extension, string, error) {
	args := c.Called(namespace, options)
	return args.Get(0).([]corev2.Extension), args.Get(1).(string), args.Error(2)
}

// RegisterExtension ...
func (c *MockClient) RegisterExtension(e *types.Extension) error {
	args := c.Called(e)
	return args.Error(0)
}

// DeregisterExtension ...
func (c *MockClient) DeregisterExtension(name, namespace string) error {
	args := c.Called(name, namespace)
	return args.Error(0)
}
