package asset

import (
	"context"
	"errors"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAssets(t *testing.T) {
	asset1 := corev2.FixtureAsset("asset1")
	asset1.URL = "https://localhost/asset1.zip"
	asset2 := corev2.FixtureAsset("asset2")
	asset2.URL = "https://localhost/asset2.zip"
	asset3 := corev2.FixtureAsset("asset3")
	asset3.URL = "https://localhost/asset3.zip"

	testCases := []struct {
		name           string
		assetList      []string
		expectedAssets []corev2.Asset
	}{
		{
			name:           "found all assets",
			assetList:      []string{"asset1", "asset2", "asset3"},
			expectedAssets: []corev2.Asset{*asset1, *asset2, *asset3},
		},
		{
			name:           "empty asset list",
			assetList:      []string{},
			expectedAssets: []corev2.Asset{},
		},
		{
			name:           "asset not found",
			assetList:      []string{"foo", "asset1"},
			expectedAssets: []corev2.Asset{*asset1},
		},
		{
			name:           "error on store",
			assetList:      []string{"bar", "asset1"},
			expectedAssets: []corev2.Asset{*asset1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a1req := storev2.NewResourceRequestFromResource(asset1)
			a2req := storev2.NewResourceRequestFromResource(asset2)
			a3req := storev2.NewResourceRequestFromResource(asset3)
			a4req := storev2.NewResourceRequestFromResource(asset3)
			a4req.Name = "foo"
			a5req := storev2.NewResourceRequestFromResource(asset3)
			a5req.Name = "bar"
			sto := &mockstore.V2MockStore{}
			cs := new(mockstore.ConfigStore)
			sto.On("GetConfigStore").Return(cs)
			cs.On("Get", mock.Anything, a1req).Return(mockstore.Wrapper[*corev2.Asset]{Value: asset1}, nil)
			cs.On("Get", mock.Anything, a2req).Return(mockstore.Wrapper[*corev2.Asset]{Value: asset2}, nil)
			cs.On("Get", mock.Anything, a3req).Return(mockstore.Wrapper[*corev2.Asset]{Value: asset3}, nil)
			cs.On("Get", mock.Anything, a4req).Return(nil, &store.ErrNotFound{})
			cs.On("Get", mock.Anything, a5req).Return(nil, errors.New("error"))

			ctx := store.NamespaceContext(context.Background(), "default")
			assets := GetAssets(ctx, sto, tc.assetList)
			assert.EqualValues(t, tc.expectedAssets, assets)
		})
	}
}
