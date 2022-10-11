package asset

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	corev2 "github.com/sensu/core/v2"
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
			var nilAsset *corev2.Asset
			store := &mockstore.MockStore{}
			store.On("GetAssetByName", mock.Anything, "asset1").Return(asset1, nil)
			store.On("GetAssetByName", mock.Anything, "asset2").Return(asset2, nil)
			store.On("GetAssetByName", mock.Anything, "asset3").Return(asset3, nil)
			store.On("GetAssetByName", mock.Anything, "foo").Return(nilAsset, nil)
			store.On("GetAssetByName", mock.Anything, "bar").Return(nilAsset, errors.New("error"))

			assets := GetAssets(context.Background(), store, tc.assetList)
			assert.EqualValues(t, tc.expectedAssets, assets)
		})
	}
}
