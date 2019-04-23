package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// DeleteAssetByName ...
func (s *MockStore) DeleteAssetByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetAssets ...
func (s *MockStore) GetAssets(ctx context.Context, pred *store.SelectionPredicate) ([]*types.Asset, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*types.Asset), args.Error(1)
}

// GetAssetByName ...
func (s *MockStore) GetAssetByName(ctx context.Context, name string) (*types.Asset, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*types.Asset), args.Error(1)
}

// UpdateAsset ...
func (s *MockStore) UpdateAsset(ctx context.Context, asset *types.Asset) error {
	args := s.Called(ctx, asset)
	return args.Error(0)
}
