package mockstore

import "github.com/sensu/sensu-go/types"

//// Assets

// GetAssets ...
func (s *MockStore) GetAssets() ([]*types.Asset, error) {
	args := s.Called()
	return args.Get(0).([]*types.Asset), args.Error(1)
}

// GetAssetByName ...
func (s *MockStore) GetAssetByName(name string) (*types.Asset, error) {
	args := s.Called(name)
	return args.Get(0).(*types.Asset), args.Error(1)
}

// DeleteAssetByName ...
func (s *MockStore) DeleteAssetByName(name string) error {
	args := s.Called(name)
	return args.Error(0)
}

// UpdateAsset ...
func (s *MockStore) UpdateAsset(asset *types.Asset) error {
	args := s.Called(asset)
	return args.Error(0)
}
