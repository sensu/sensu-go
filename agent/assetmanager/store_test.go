package assetmanager

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

type storeTest struct {
	store      *AssetStore
	newAssetFn func(*types.Asset) *RuntimeAsset
	newSetFn   func([]*RuntimeAsset) *RuntimeAssetSet
}

func newStoreTest() *storeTest {
	return &storeTest{
		store: NewAssetStore(),
		newAssetFn: func(a *types.Asset) *RuntimeAsset {
			return NewRuntimeAsset(a, "")
		},
		newSetFn: func(a []*RuntimeAsset) *RuntimeAssetSet {
			return NewRuntimeAssetSet(a, []string{})
		},
	}
}

func TestNew(t *testing.T) {
	store := NewAssetStore()
	assert.NotNil(t, store)
	assert.Empty(t, store.assets)
	assert.Empty(t, store.assetSets)
	assert.NotNil(t, store.rwMutex)
}

func TestFetchAsset(t *testing.T) {
	asset := types.FixtureAsset("name")
	test := newStoreTest()
	runtimeAsset := test.store.FetchAsset(asset, test.newAssetFn)
	assert.NotNil(t, runtimeAsset)

	secondRun := test.store.FetchAsset(asset, test.newAssetFn)
	assert.Equal(t, &runtimeAsset, &secondRun)
}

func TestFetchSet(t *testing.T) {
	asset := types.FixtureAsset("name")
	runtimeAssets := []*RuntimeAsset{&RuntimeAsset{asset: asset}}
	test := newStoreTest()
	assetSet := test.store.FetchSet(runtimeAssets, test.newSetFn)
	assert.NotNil(t, assetSet)

	secondRun := test.store.FetchSet(runtimeAssets, test.newSetFn)
	assert.Equal(t, &assetSet, &secondRun)
}

func TestFetchClear(t *testing.T) {
	test := newStoreTest()
	test.store.Clear()
	assert.Empty(t, test.store.assets)
	assert.Empty(t, test.store.assetSets)
}
