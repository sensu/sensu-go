package testing

import (
	corev2 "github.com/sensu/core/v2"
)

// FetchEntity for use with mock lib
func (c *MockClient) FetchEntity(ID string) (*corev2.Entity, error) {
	args := c.Called(ID)
	return args.Get(0).(*corev2.Entity), args.Error(1)
}

// DeleteEntity for use with mock lib
func (c *MockClient) DeleteEntity(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}

// UpdateEntity for use with mock lib
func (c *MockClient) UpdateEntity(entity *corev2.Entity) error {
	args := c.Called(entity)
	return args.Error(0)
}

// CreateEntity for use with mock lib
func (c *MockClient) CreateEntity(entity *corev2.Entity) error {
	args := c.Called(entity)
	return args.Error(0)
}
