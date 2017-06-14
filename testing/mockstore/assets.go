package mockstore

import "github.com/sensu/sensu-go/types"

// GetAssets ...
func (s *MockStore) GetAssets(org string) ([]*types.Asset, error) {
	args := s.Called(org)
	return args.Get(0).([]*types.Asset), args.Error(1)
}

// GetAssetByName ...
func (s *MockStore) GetAssetByName(org, name string) (*types.Asset, error) {
	args := s.Called(org, name)
	return args.Get(0).(*types.Asset), args.Error(1)
}

// DeleteAssetByName ...
func (s *MockStore) DeleteAssetByName(org, name string) error {
	args := s.Called(name)
	return args.Error(0)
}

// UpdateAsset ...
func (s *MockStore) UpdateAsset(asset *types.Asset) error {
	args := s.Called(asset)
	return args.Error(0)
}
