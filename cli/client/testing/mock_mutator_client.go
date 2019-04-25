package testing

import (
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// CreateMutator for use with mock package
func (c *MockClient) CreateMutator(m *types.Mutator) error {
	args := c.Called(m)
	return args.Error(0)
}

// DeleteMutator for use with mock package
func (c *MockClient) DeleteMutator(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}

// FetchMutator for use with mock package
func (c *MockClient) FetchMutator(name string) (*types.Mutator, error) {
	args := c.Called(name)
	return args.Get(0).(*types.Mutator), args.Error(1)
}

// UpdateMutator for use with mock package
func (c *MockClient) UpdateMutator(m *types.Mutator) error {
	args := c.Called(m)
	return args.Error(0)
}

// ListMutators for use with mock lib
func (c *MockClient) ListMutators(namespace string, options client.ListOptions) ([]corev2.Mutator, error) {
	args := c.Called(namespace, options)
	return args.Get(0).([]corev2.Mutator), args.Error(1)
}
