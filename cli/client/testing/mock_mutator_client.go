package testing

import (
	corev2 "github.com/sensu/core/v2"
)

// CreateMutator for use with mock package
func (c *MockClient) CreateMutator(m *corev2.Mutator) error {
	args := c.Called(m)
	return args.Error(0)
}

// DeleteMutator for use with mock package
func (c *MockClient) DeleteMutator(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}

// FetchMutator for use with mock package
func (c *MockClient) FetchMutator(name string) (*corev2.Mutator, error) {
	args := c.Called(name)
	return args.Get(0).(*corev2.Mutator), args.Error(1)
}

// UpdateMutator for use with mock package
func (c *MockClient) UpdateMutator(m *corev2.Mutator) error {
	args := c.Called(m)
	return args.Error(0)
}
