package testing

import "github.com/sensu/sensu-go/types"

// ListAssets for use with mock lib
func (c *MockClient) ListAssets(namespace string) ([]types.Asset, error) {
	args := c.Called(namespace)
	return args.Get(0).([]types.Asset), args.Error(1)
}

// FetchAsset for use with mock lib
func (c *MockClient) FetchAsset(name string) (*types.Asset, error) {
	args := c.Called(name)
	return args.Get(0).(*types.Asset), args.Error(1)
}

// CreateAsset for use with mock lib
func (c *MockClient) CreateAsset(asset *types.Asset) error {
	args := c.Called(asset)
	return args.Error(0)
}

// UpdateAsset for use with mock lib
func (c *MockClient) UpdateAsset(asset *types.Asset) error {
	args := c.Called(asset)
	return args.Error(0)
}
