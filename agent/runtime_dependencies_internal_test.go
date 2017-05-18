package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/suite"
)

type ManagerTestSuite struct {
	suite.Suite
	agent       *Agent
	dep         *runtimeDependency
	manager     *DependencyManager
	assetServer *httptest.Server
}

func (suite *ManagerTestSuite) SetupTest() {
	// Ex script
	exBody := "abc"
	exHash := stringToSHA256(exBody)

	// Setup a fake server to fake retrieving the asset
	suite.assetServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text")
		fmt.Fprintf(w, exBody)
	}))

	// Ex. asset
	asset := &types.Asset{
		Name: "ruby24",
		Hash: exHash,
		URL:  suite.assetServer.URL + "/myfile",
	}

	// Create a fake cache directory so that we have a safe place to test results
	tmpDir, _ := ioutil.TempDir(os.TempDir(), "agent-deps-")

	// Ex. manager
	manager := NewDependencyManager(tmpDir)
	dep := &runtimeDependency{manager: manager, asset: asset}
	manager.knownDeps["test"] = dep

	suite.dep = dep
	suite.manager = manager
}

func (suite *ManagerTestSuite) AfterTest() {
	// Shutdown asset server
	suite.assetServer.Close()

	// Remove tmpdir
	os.RemoveAll(suite.manager.cacheDir)
}

func (suite *ManagerTestSuite) TestNewDepManager() {
	manager := NewDependencyManager("./tmp")

	suite.NotNil(manager)
	suite.Contains(manager.cacheDir, "tmp")
	suite.Contains(manager.cacheDir, "deps")
	suite.Empty(manager.knownDeps)
}

func (suite *ManagerTestSuite) TestManagerPaths() {
	paths := suite.manager.paths()
	suite.NotEmpty(paths)
	suite.Contains(paths, suite.dep.path())
}

func (suite *ManagerTestSuite) TestManagerEnv() {
	suite.manager.BaseEnv = []string{"PATH=/usr/bin"}

	// Not memoized
	resEnv := suite.manager.Env()
	suite.NotEmpty(resEnv)
	suite.NotEmpty(resEnv[0])
	suite.Contains(resEnv[0], suite.manager.cacheDir)

	// memoized
	suite.manager.env = []string{"PATH=test"}
	resEnv = suite.manager.Env()
	suite.NotEmpty(resEnv)
	suite.Equal("PATH=test", resEnv[0])

	// Other vars
	suite.manager.env = nil
	suite.manager.BaseEnv = []string{
		"PATH=tmp/bin",
		"LD_LIBRARY_PATH=tmp/lib",
		"CPATH=tmp/include",
		"CAKES=are.tasty",
	}
	resEnv = suite.manager.Env()
	suite.NotEmpty(resEnv)
	suite.Contains(resEnv[0], suite.manager.cacheDir)
	suite.Contains(resEnv[1], suite.manager.cacheDir)
	suite.Contains(resEnv[2], suite.manager.cacheDir)
	suite.NotContains(resEnv[3], suite.manager.cacheDir)
}

func (suite *ManagerTestSuite) TestManagerInstall() {
	err := suite.manager.Install()
	suite.NoError(err)
}

func (suite *ManagerTestSuite) TestManagerReset() {
	suite.manager.Reset()
	suite.Nil(suite.manager.env)
	suite.Empty(suite.manager.knownDeps)
	suite.NotEmpty(suite.manager.BaseEnv)
}

func (suite *ManagerTestSuite) TestManagerSetConfig() {
	suite.manager.SetCacheDir("./tmp")
	suite.Nil(suite.manager.env)
	suite.NotEmpty(suite.manager.cacheDir)
}

type DependencyTestSuite struct {
	suite.Suite

	assetServer  *httptest.Server
	dep          *runtimeDependency
	manager      *DependencyManager
	responseBody string
	responseType string
}

func (suite *DependencyTestSuite) SetupTest() {
	// Setup a fake server to fake retrieving the asset
	suite.assetServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", suite.responseType)
		fmt.Fprintf(w, suite.responseBody)
	}))

	// Default response
	suite.responseBody = ""
	suite.responseType = "text"

	// Create a fake cache directory so that we have a safe place to test results
	tmpDir, _ := ioutil.TempDir(os.TempDir(), "agent-deps-test")

	// Ex. Dep
	suite.manager = &DependencyManager{cacheDir: tmpDir}
	suite.dep = &runtimeDependency{
		manager: suite.manager,
		asset: &types.Asset{
			Name: "ruby24",
			Hash: "123456",
			URL:  suite.assetServer.URL + "/myfile",
		},
	}
}

func (suite *DependencyTestSuite) AfterTest() {
	// Shutdown asset server
	suite.assetServer.Close()

	// Remove tmpdir
	os.RemoveAll(suite.dep.manager.cacheDir)
}

func (suite *DependencyTestSuite) TestFetch() {
	suite.responseBody = "abc"

	res, err := suite.dep.fetch()
	suite.NotNil(res)
	suite.NoError(err)
}

func (suite *DependencyTestSuite) TestInstall() {
	suite.responseBody = "abc"
	suite.dep.asset.Hash = stringToSHA256(suite.responseBody)

	err := suite.dep.install()
	suite.NoError(err)
}

func (suite *DependencyTestSuite) TestInstallBadAssetHash() {
	suite.responseBody = "abc"
	suite.dep.asset.Hash = "bad bad hash boy"

	err := suite.dep.install()
	suite.Error(err)
}

func (suite *DependencyTestSuite) TestIsCached() {
	fmt.Println(suite.dep.path())
	cached, err := suite.dep.isCached()
	suite.False(cached)
	suite.NoError(err)

	os.MkdirAll(suite.dep.path(), 0755)
	cached, err = suite.dep.isCached()
	suite.True(cached)
	suite.NoError(err)
}

func (suite *DependencyTestSuite) TestIsCachedDirIsNotDirectory() {
	os.MkdirAll(path.Dir(suite.dep.path()), 0755)
	os.OpenFile(suite.dep.path(), os.O_RDONLY|os.O_CREATE, 0666)

	cached, err := suite.dep.isCached()
	suite.True(cached)
	suite.Error(err)
}

func TestRuntimeDependencies(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
	suite.Run(t, new(DependencyTestSuite))
}

//
// Helpers

func stringToSHA256(hash string) string {
	h := sha256.New()
	h.Write([]byte(hash))
	return hex.EncodeToString(h.Sum(nil))
}
