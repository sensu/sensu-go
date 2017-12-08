package assetmanager

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/suite"
)

type ManagerTestSuite struct {
	suite.Suite
	cacheDir string
	manager  *Manager
}

func (suite *ManagerTestSuite) SetupTest() {
	// Create a fake cache directory so that we have a safe place to test results
	tmpDir, _ := ioutil.TempDir(os.TempDir(), "agent-deps-")
	suite.cacheDir = tmpDir

	// Ex. manager
	manager := &Manager{}
	manager.entity = &types.Entity{}
	manager.store = NewAssetStore()
	manager.factory = &AssetFactory{
		CacheDir: tmpDir,
		BaseEnv:  os.Environ(),
	}
	suite.manager = manager
}

func (suite *ManagerTestSuite) AfterTest() {
	// Remove tmpdir
	suite.NoError(os.RemoveAll(suite.cacheDir))
}

func (suite *ManagerTestSuite) TestNewManager() {
	manager := New("./tmp", &types.Entity{})

	suite.NotNil(manager)
	suite.NotNil(manager.store)
	suite.NotNil(manager.factory)
}

func (suite *ManagerTestSuite) TestSetCacheDir() {
	asset := NewRuntimeAsset(types.FixtureAsset("asset"), "")
	suite.manager.store.assets["123"] = asset

	suite.manager.SetCacheDir("my-test-dir")
	suite.Contains(suite.manager.factory.CacheDir, "my-test-dir")
	suite.Empty(suite.manager.store.assets, "clears existing assets from store")
}

func (suite *ManagerTestSuite) TestReset() {
	asset := NewRuntimeAsset(types.FixtureAsset("asset"), "")
	suite.manager.store.assets["123"] = asset

	suite.manager.Reset()
	suite.Empty(suite.manager.store.assets, "clears existing assets from store")
}

func (suite *ManagerTestSuite) TestRegisterSet() {
	assets := []types.Asset{*types.FixtureAsset("asset")}
	assetSet := suite.manager.RegisterSet(assets)
	suite.NotEmpty(assetSet.Env)
	suite.NotEmpty(assetSet.assets)

	store := suite.manager.store
	suite.NotEmpty(store.assets)
	suite.NotEmpty(store.assetSets)
}

func TestManager(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}
