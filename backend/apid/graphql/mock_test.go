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

type MockCheckClient struct {
	mock.Mock
}

func (c *MockCheckClient) CreateCheck(ctx context.Context, check *corev2.CheckConfig) error {
	return c.Called(ctx, check).Error(0)
}

func (c *MockCheckClient) UpdateCheck(ctx context.Context, check *corev2.CheckConfig) error {
	return c.Called(ctx, check).Error(0)
}

func (c *MockCheckClient) DeleteCheck(ctx context.Context, name string) error {
	return c.Called(ctx, name).Error(0)
}

func (c *MockCheckClient) ExecuteCheck(ctx context.Context, name string, req *corev2.AdhocRequest) error {
	return c.Called(ctx, name, req).Error(0)
}

func (c *MockCheckClient) FetchCheck(ctx context.Context, name string) (*corev2.CheckConfig, error) {
	args := c.Called(ctx, name)
	return args.Get(0).(*corev2.CheckConfig), args.Error(1)
}

func (c *MockCheckClient) ListChecks(ctx context.Context) ([]*corev2.CheckConfig, error) {
	args := c.Called(ctx)
	return args.Get(0).([]*corev2.CheckConfig), args.Error(1)
}
