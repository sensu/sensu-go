package mockstore

import (
	"context"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// DeleteAssetByName ...
func (s *MockStore) DeleteAssetByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetAssets ...
func (s *MockStore) GetAssets(ctx context.Context, pred *store.SelectionPredicate) ([]*v2.Asset, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*v2.Asset), args.Error(1)
}

// GetAssetByName ...
func (s *MockStore) GetAssetByName(ctx context.Context, name string) (*v2.Asset, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*v2.Asset), args.Error(1)
}

// UpdateAsset ...
func (s *MockStore) UpdateAsset(ctx context.Context, asset *v2.Asset) error {
	args := s.Called(ctx, asset)
	return args.Error(0)
}
