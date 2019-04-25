package testing

import (
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// CreateSilenced for use with mock lib
func (c *MockClient) CreateSilenced(silenced *types.Silenced) error {
	args := c.Called(silenced)
	return args.Error(0)
}

// UpdateSilenced for use with mock lib
func (c *MockClient) UpdateSilenced(silenced *types.Silenced) error {
	args := c.Called(silenced)
	return args.Error(0)
}

// DeleteSilenced for use with mock lib
func (c *MockClient) DeleteSilenced(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}

// FetchSilenced for use with mock lib
func (c *MockClient) FetchSilenced(id string) (*types.Silenced, error) {
	args := c.Called(id)
	return args.Get(0).(*types.Silenced), args.Error(1)
}

// ListSilenceds for use with mock lib
func (c *MockClient) ListSilenceds(namespace, sub, check string, options client.ListOptions) ([]corev2.Silenced, error) {
	args := c.Called(namespace, sub, check, options)
	return args.Get(0).([]corev2.Silenced), args.Error(1)
}
