package assetmanager

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type managerTest struct {
	cacheDir string
	manager  *Manager
}

func (m *managerTest) Dispose(t *testing.T) {
	require.NoError(t, os.RemoveAll(m.cacheDir))
}

func newManagerTest(t *testing.T) *managerTest {
	// Create a fake cache directory so that we have a safe place to test results
	test := &managerTest{}
	tmpDir, err := ioutil.TempDir(os.TempDir(), "agent-deps-")
	require.NoError(t, err)
	test.cacheDir = tmpDir

	// Ex. manager
	manager := &Manager{}
	manager.entity = &types.Entity{}
	manager.store = NewAssetStore()
	manager.factory = &AssetFactory{
		CacheDir: tmpDir,
		BaseEnv:  os.Environ(),
	}
	test.manager = manager
	return test
}

func TestNewManager(t *testing.T) {
	manager := New("./tmp", &types.Entity{})

	require.NotNil(t, manager)
	require.NotNil(t, manager.store)
	require.NotNil(t, manager.factory)
}

func TestSetCacheDir(t *testing.T) {
	test := newManagerTest(t)
	defer test.Dispose(t)
	asset := NewRuntimeAsset(types.FixtureAsset("asset"), "")
	test.manager.store.assets["123"] = asset

	test.manager.SetCacheDir("my-test-dir")
	assert.Contains(t, test.manager.factory.CacheDir, "my-test-dir")
	assert.Empty(t, test.manager.store.assets, "clears existing assets from store")
}

func TestReset(t *testing.T) {
	test := newManagerTest(t)
	defer test.Dispose(t)
	asset := NewRuntimeAsset(types.FixtureAsset("asset"), "")
	test.manager.store.assets["123"] = asset

	test.manager.Reset()
	require.Empty(t, test.manager.store.assets, "clears existing assets from store")
}

func TestRegisterSet(t *testing.T) {
	test := newManagerTest(t)
	defer test.Dispose(t)
	assets := []types.Asset{*types.FixtureAsset("asset")}
	assetSet := test.manager.RegisterSet(assets)
	assert.NotEmpty(t, assetSet.Env)
	assert.NotEmpty(t, assetSet.assets)

	store := test.manager.store
	assert.NotEmpty(t, store.assets)
	assert.NotEmpty(t, store.assetSets)
}
