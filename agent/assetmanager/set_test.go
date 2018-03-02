package assetmanager

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type assetTest struct {
	cacheDir string
	asset    *RuntimeAsset
	assetSet *RuntimeAssetSet
}

func (a *assetTest) Dispose(t *testing.T) {
	// Remove tmpdir
	_ = os.RemoveAll(a.cacheDir)
}

func newAssetTest(t *testing.T) (*httptest.Server, *assetTest) {
	test := &assetTest{}
	exBody := readFixture("rubby-on-rails.tar")
	exSha512 := stringToSHA512(exBody)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/plain")
		fmt.Fprintf(w, exBody)
	}))
	// Ex script

	// Ex. asset
	asset := types.FixtureAsset("asset")
	asset.Name = "ruby24"
	asset.Sha512 = exSha512
	asset.URL = server.URL + "/myfile"

	// Create a fake cache directory so that we have a safe place to test results
	tmpDir, err := ioutil.TempDir(os.TempDir(), fmt.Sprintf("agent-deps-%d", time.Now().UnixNano()))
	require.NoError(t, err)
	test.cacheDir = tmpDir

	// Ex. set
	runtimeAsset := NewRuntimeAsset(asset, tmpDir)
	runtimeSet := NewRuntimeAssetSet([]*RuntimeAsset{runtimeAsset}, []string{})
	test.asset = runtimeAsset
	test.assetSet = runtimeSet

	return server, test
}

func TestNewSet(t *testing.T) {
	asset := NewRuntimeAsset(types.FixtureAsset("asset"), "test")
	set := NewRuntimeAssetSet([]*RuntimeAsset{asset}, []string{"TEST="})

	assert.NotNil(t, set)
	assert.NotEmpty(t, set.env)
	assert.NotEmpty(t, set.assets)
}

func TestManagerPaths(t *testing.T) {
	server, test := newAssetTest(t)
	defer server.Close()
	defer test.Dispose(t)

	paths := test.assetSet.paths()
	assert.NotEmpty(t, paths)
	assert.Contains(t, paths, test.asset.path)
}

func TestComputeEnv(t *testing.T) {
	server, test := newAssetTest(t)
	defer server.Close()
	defer test.Dispose(t)

	test.asset.path = filepath.Join("tmp", "test")
	test.assetSet.computeEnv([]string{
		"PATH=tmp/bin",
		"LD_LIBRARY_PATH=tmp/lib",
		"CPATH=tmp/include",
		"CAKES=are.tasty",
	})

	resEnv := test.assetSet.Env()
	assert.NotEmpty(t, resEnv)
	assert.Contains(t, resEnv[0], test.asset.path)
	assert.Contains(t, resEnv[1], test.asset.path)
	assert.Contains(t, resEnv[2], test.asset.path)
	assert.NotContains(t, resEnv[3], test.asset.path)
}

func TestManagerInstall(t *testing.T) {
	server, test := newAssetTest(t)
	defer server.Close()
	defer test.Dispose(t)

	require.NoError(t, test.assetSet.InstallAll())
}
