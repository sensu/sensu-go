package testing

import (
	corev2 "github.com/sensu/core/v2"
)

// FetchAsset for use with mock lib
func (c *MockClient) FetchAsset(name string) (*corev2.Asset, error) {
	args := c.Called(name)
	return args.Get(0).(*corev2.Asset), args.Error(1)
}

// CreateAsset for use with mock lib
func (c *MockClient) CreateAsset(asset *corev2.Asset) error {
	args := c.Called(asset)
	return args.Error(0)
}

// UpdateAsset for use with mock lib
func (c *MockClient) UpdateAsset(asset *corev2.Asset) error {
	args := c.Called(asset)
	return args.Error(0)
}
