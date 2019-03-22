package testing

import "github.com/sensu/sensu-go/types"

// ListEntities for use with mock lib
func (c *MockClient) ListEntities(namespace string) ([]types.Entity, error) {
	args := c.Called(namespace)
	return args.Get(0).([]types.Entity), args.Error(1)
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
