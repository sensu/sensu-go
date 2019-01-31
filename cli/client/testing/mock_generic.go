package testing

import "github.com/sensu/sensu-go/types"

// Delete ...
func (c *MockClient) Delete(path string) error {
	args := c.Called(path)
	return args.Error(0)
}

// Get ...
func (c *MockClient) Get(path string, obj interface{}) error {
	args := c.Called(path, obj)
	return args.Error(0)
}

// List ...
func (c *MockClient) List(path string, objs interface{}) error {
	args := c.Called(path, objs)
	return args.Error(0)
}

// Post ...
func (c *MockClient) Post(path string, obj interface{}) error {
	args := c.Called(path, obj)
	return args.Error(0)
}

// PutResource ...
func (c *MockClient) PutResource(r types.Resource) error {
	args := c.Called(r)
	return args.Error(0)
}
