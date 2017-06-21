package assetmanager

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/suite"
)

type AssetSetTestSuite struct {
	suite.Suite

	cacheDir    string
	asset       *RuntimeAsset
	assetSet    *RuntimeAssetSet
	assetServer *httptest.Server
}

func (suite *AssetSetTestSuite) SetupTest() {
	// Ex script
	exBody := readFixture("rubby-on-rails.tar")
	exSha512 := stringToSHA512(exBody)

	// Setup a fake server to fake retrieving the asset
	suite.assetServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text")
		fmt.Fprintf(w, exBody)
	}))

	// Ex. asset
	asset := types.FixtureAsset("asset")
	asset.Name = "ruby24"
	asset.Sha512 = exSha512
	asset.URL = suite.assetServer.URL + "/myfile"

	// Create a fake cache directory so that we have a safe place to test results
	tmpDir, _ := ioutil.TempDir(os.TempDir(), "agent-deps-")
	suite.cacheDir = tmpDir

	// Ex. set
	runtimeAsset := NewRuntimeAsset(asset, tmpDir)
	runtimeSet := NewRuntimeAssetSet([]*RuntimeAsset{runtimeAsset}, []string{})
	suite.asset = runtimeAsset
	suite.assetSet = runtimeSet
}

func (suite *AssetSetTestSuite) AfterTest() {
	// Shutdown asset server
	suite.assetServer.Close()

	// Remove tmpdir
	os.RemoveAll(suite.cacheDir)
}

func (suite *AssetSetTestSuite) TestNewSet() {
	asset := NewRuntimeAsset(types.FixtureAsset("asset"), "test")
	set := NewRuntimeAssetSet([]*RuntimeAsset{asset}, []string{"TEST="})

	suite.NotNil(set)
	suite.NotEmpty(set.env)
	suite.NotEmpty(set.assets)
}

func (suite *AssetSetTestSuite) TestManagerPaths() {
	paths := suite.assetSet.paths()
	suite.NotEmpty(paths)
	suite.Contains(paths, suite.asset.path)
}

func (suite *AssetSetTestSuite) TestComputeEnv() {
	suite.asset.path = filepath.Join("tmp", "test")
	suite.assetSet.computeEnv([]string{
		"PATH=tmp/bin",
		"LD_LIBRARY_PATH=tmp/lib",
		"CPATH=tmp/include",
		"CAKES=are.tasty",
	})

	resEnv := suite.assetSet.Env()
	suite.NotEmpty(resEnv)
	suite.Contains(resEnv[0], suite.asset.path)
	suite.Contains(resEnv[1], suite.asset.path)
	suite.Contains(resEnv[2], suite.asset.path)
	suite.NotContains(resEnv[3], suite.asset.path)
}

func (suite *AssetSetTestSuite) TestManagerInstall() {
	err := suite.assetSet.InstallAll()
	suite.NoError(err)
}

func TestRuntimeAssetSet(t *testing.T) {
	suite.Run(t, new(AssetSetTestSuite))
}
