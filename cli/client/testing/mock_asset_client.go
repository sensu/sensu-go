package testing

import (
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// ListAssets for use with mock lib
func (c *MockClient) ListAssets(namespace string, options client.ListOptions) ([]corev2.Asset, string, error) {
	args := c.Called(namespace, options)
	return args.Get(0).([]corev2.Asset), args.Get(1).(string), args.Error(2)
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
