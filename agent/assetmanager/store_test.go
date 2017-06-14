package assetmanager

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/suite"
)

type StoreTestSuite struct {
	suite.Suite
	store      *AssetStore
	newAssetFn func(*types.Asset) *RuntimeAsset
	newSetFn   func([]*RuntimeAsset) *RuntimeAssetSet
}

func (suite *StoreTestSuite) SetupTest() {
	suite.store = NewAssetStore()
	suite.newAssetFn = func(a *types.Asset) *RuntimeAsset {
		return NewRuntimeAsset(a, "")
	}
	suite.newSetFn = func(a []*RuntimeAsset) *RuntimeAssetSet {
		return NewRuntimeAssetSet(a, []string{})
	}
}

func (suite *StoreTestSuite) TestNew() {
	store := NewAssetStore()
	suite.NotNil(store)
	suite.Empty(store.assets)
	suite.Empty(store.assetSets)
	suite.NotNil(store.rwMutex)
}

func (suite *StoreTestSuite) TestFetchAsset() {
	asset := types.FixtureAsset("name")
	runtimeAsset := suite.store.FetchAsset(asset, suite.newAssetFn)
	suite.NotNil(runtimeAsset)

	secondRun := suite.store.FetchAsset(asset, suite.newAssetFn)
	suite.Equal(&runtimeAsset, &secondRun)
}

func (suite *StoreTestSuite) TestFetchSet() {
	asset := types.FixtureAsset("name")
	runtimeAssets := []*RuntimeAsset{&RuntimeAsset{asset: asset}}
	assetSet := suite.store.FetchSet(runtimeAssets, suite.newSetFn)
	suite.NotNil(assetSet)

	secondRun := suite.store.FetchSet(runtimeAssets, suite.newSetFn)
	suite.Equal(&assetSet, &secondRun)
}

func (suite *StoreTestSuite) TestFetchClear() {
	suite.store.Clear()
	suite.Empty(suite.store.assets)
	suite.Empty(suite.store.assetSets)
}

func TestStore(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}
