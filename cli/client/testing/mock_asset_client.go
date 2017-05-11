package testing

import "github.com/sensu/sensu-go/types"

// ListAssets for use with mock lib
func (c *MockClient) ListAssets() ([]types.Asset, error) {
	args := c.Called()
	return args.Get(0).([]types.Asset), args.Error(1)
}

// CreateAsset for use with mock lib
func (c *MockClient) CreateAsset(asset *types.Asset) error {
	args := c.Called(asset)
	return args.Error(0)
}
