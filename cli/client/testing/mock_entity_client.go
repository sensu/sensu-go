package testing

import (
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// ListEntities for use with mock lib
func (c *MockClient) ListEntities(namespace string, options client.ListOptions) ([]corev2.Entity, string, error) {
	args := c.Called(namespace, options)
	return args.Get(0).([]corev2.Entity), args.Get(1).(string), args.Error(2)
}

// FetchEntity for use with mock lib
func (c *MockClient) FetchEntity(ID string) (*types.Entity, error) {
	args := c.Called(ID)
	return args.Get(0).(*types.Entity), args.Error(1)
}

// DeleteEntity for use with mock lib
func (c *MockClient) DeleteEntity(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}

// UpdateEntity for use with mock lib
func (c *MockClient) UpdateEntity(entity *types.Entity) error {
	args := c.Called(entity)
	return args.Error(0)
}

// CreateEntity for use with mock lib
func (c *MockClient) CreateEntity(entity *types.Entity) error {
	args := c.Called(entity)
	return args.Error(0)
}
