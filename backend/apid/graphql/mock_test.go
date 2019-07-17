package graphql

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/mock"
)

type MockAssetClient struct {
	mock.Mock
}

func (a *MockAssetClient) ListAssets(ctx context.Context) ([]*corev2.Asset, error) {
	args := a.Called(ctx)
	return args.Get(0).([]*corev2.Asset), args.Error(1)
}

func (a *MockAssetClient) FetchAsset(ctx context.Context, name string) (*corev2.Asset, error) {
	args := a.Called(ctx, name)
	return args.Get(0).(*corev2.Asset), args.Error(1)
}

func (a *MockAssetClient) CreateAsset(ctx context.Context, asset *corev2.Asset) error {
	return a.Called(ctx, asset).Error(0)
}

func (a *MockAssetClient) UpdateAsset(ctx context.Context, asset *corev2.Asset) error {
	return a.Called(ctx, asset).Error(0)
}
