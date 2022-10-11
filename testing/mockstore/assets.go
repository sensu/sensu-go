package mockstore

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// DeleteAssetByName ...
func (s *MockStore) DeleteAssetByName(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// GetAssets ...
func (s *MockStore) GetAssets(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Asset, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*corev2.Asset), args.Error(1)
}

// GetAssetByName ...
func (s *MockStore) GetAssetByName(ctx context.Context, name string) (*corev2.Asset, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*corev2.Asset), args.Error(1)
}

// UpdateAsset ...
func (s *MockStore) UpdateAsset(ctx context.Context, asset *corev2.Asset) error {
	args := s.Called(ctx, asset)
	return args.Error(0)
}
