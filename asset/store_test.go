package asset

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAssets(t *testing.T) {
	asset1 := types.FixtureAsset("asset1")
	asset1.URL = "https://localhost/asset1.zip"
	asset2 := types.FixtureAsset("asset2")
	asset2.URL = "https://localhost/asset2.zip"
	asset3 := types.FixtureAsset("asset3")
	asset3.URL = "https://localhost/asset3.zip"

	testCases := []struct {
		name           string
		assetList      []string
		expectedAssets []types.Asset
	}{
		{
			name:           "found all assets",
			assetList:      []string{"asset1", "asset2", "asset3"},
			expectedAssets: []types.Asset{*asset1, *asset2, *asset3},
		},
		{
			name:           "empty asset list",
			assetList:      []string{},
			expectedAssets: []types.Asset{},
		},
		{
			name:           "asset not found",
			assetList:      []string{"foo", "asset1"},
			expectedAssets: []types.Asset{*asset1},
		},
		{
			name:           "error on store",
			assetList:      []string{"bar", "asset1"},
			expectedAssets: []types.Asset{*asset1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var nilAsset *types.Asset
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
