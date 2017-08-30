package testing

import "github.com/sensu/sensu-go/types"

// CreateOrganization for use with mock lib
func (c *MockClient) CreateOrganization(org *types.Organization) error {
	args := c.Called(org)
	return args.Error(0)
}

// DeleteOrganization for use with mock lib
func (c *MockClient) DeleteOrganization(org string) error {
	args := c.Called(org)
	return args.Error(0)
}

// ListOrganizations for use with mock lib
func (c *MockClient) ListOrganizations() ([]types.Organization, error) {
	args := c.Called()
	return args.Get(0).([]types.Organization), args.Error(1)
}

// FetchOrganization for use with mock lib
func (c *MockClient) FetchOrganization(org string) (*types.Organization, error) {
	args := c.Called(org)
	return args.Get(0).(*types.Organization), args.Error(1)
}
