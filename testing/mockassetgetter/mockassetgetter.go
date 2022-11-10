package mockassetgetter

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/stretchr/testify/mock"
)

// MockAssetGetter ...
type MockAssetGetter struct {
	mock.Mock
}

// Get ...
func (m *MockAssetGetter) Get(ctx context.Context, a *corev2.Asset) (*asset.RuntimeAsset, error) {
	args := m.Called(ctx, a)
	return args.Get(0).(*asset.RuntimeAsset), args.Error(1)
}
