package testing

import "github.com/sensu/sensu-go/types"

// ListEntities for use with mock lib
func (c *MockClient) ListEntities() ([]types.Entity, error) {
	args := c.Called()
	return args.Get(0).([]types.Entity), args.Error(1)
}

// FetchEntity for use with mock lib
func (c *MockClient) FetchEntity(ID string) (*types.Entity, error) {
	args := c.Called(ID)
	return args.Get(0).(*types.Entity), args.Error(1)
}

// DeleteEntity for use with mock lib
func (c *MockClient) DeleteEntity(entity *types.Entity) error {
	args := c.Called(entity)
	return args.Error(0)
}
