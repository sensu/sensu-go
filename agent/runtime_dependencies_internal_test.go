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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ManagerTestSuite struct {
	suite.Suite
	agent       *Agent
	asset       *types.Asset
	manager     *dependencyManager
	assetServer *httptest.Server
}

func (e *ManagerTestSuite) SetupTest() {
	// Ex script
	exBody := "abc"
	exHash := stringToSHA256(exBody)

	// Setup a fake server to fake retrieving the asset
	e.assetServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text")
		fmt.Fprintf(w, exBody)
	}))

	// Create a fake cache directory so that we have a safe place to test results
	tmpDir, _ := ioutil.TempDir(os.TempDir(), "agent-deps-")
	agent := &Agent{config: &Config{CacheDir: tmpDir}}

	asset := &types.Asset{
		Name: "ruby24",
		Hash: exHash,
		URL:  e.assetServer.URL + "/myfile",
	}
	manager := &dependencyManager{
		agent: agent,
		dependencies: []*runtimeDependency{
			{agent: agent, asset: asset},
		},
	}

	e.manager = manager
	e.agent = agent
	e.asset = asset
}

func (e *ManagerTestSuite) AfterTest() {
	// Shutdown asset server
	e.assetServer.Close()

	// Remove tmpdir
	os.RemoveAll(e.agent.config.CacheDir)
}

func (e *ManagerTestSuite) TestNewDepManager() {
	check := types.FixtureCheck("test")
	check.RuntimeDependencies = append(check.RuntimeDependencies, *e.asset)

	manager := newDependencyManager(e.agent, check)
	assert.NotNil(e.T(), manager)
	assert.NotEmpty(e.T(), manager.dependencies)
}

func (e *ManagerTestSuite) TestManagerPaths() {
	paths := e.manager.paths()
	assert.NotEmpty(e.T(), paths)
	assert.NotEmpty(e.T(), paths[0])
}

func (e *ManagerTestSuite) TestManagerInject() {
	testEnv := []string{"PATH=/usr/bin"}
	resEnv := e.manager.injectIntoEnv(testEnv)

	assert.NotEmpty(e.T(), resEnv)
	assert.NotEmpty(e.T(), resEnv[0])
	assert.Regexp(e.T(), "^PATH=.*;/usr/bin", resEnv[0])
}

func (e *ManagerTestSuite) TestManagerInstall() {
	err := e.manager.install()
	assert.NoError(e.T(), err)
}

type DependencyTestSuite struct {
	suite.Suite

	assetServer  *httptest.Server
	dep          *runtimeDependency
	responseBody string
	responseType string
}

func (e *DependencyTestSuite) SetupTest() {
	// Setup a fake server to fake retrieving the asset
	e.assetServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", e.responseType)
		fmt.Fprintf(w, e.responseBody)
	}))

	// Create a fake cache directory so that we have a safe place to test results
	tmpDir, _ := ioutil.TempDir(os.TempDir(), "agent-deps-test")

	e.responseBody = ""
	e.responseType = "text"
	e.dep = &runtimeDependency{
		agent: &Agent{config: &Config{CacheDir: tmpDir}},
		asset: &types.Asset{Name: "ruby24", URL: e.assetServer.URL + "/myfile"},
	}
}

func (e *DependencyTestSuite) AfterTest() {
	// Shutdown asset server
	e.assetServer.Close()

	// Remove tmpdir
	os.RemoveAll(e.dep.agent.config.CacheDir)
}

func (e *DependencyTestSuite) TestFetch() {
	e.responseBody = "abc"

	res, err := e.dep.fetch()
	assert.NotNil(e.T(), res)
	assert.NoError(e.T(), err)
}

func (e *DependencyTestSuite) TestInstall() {
	e.responseBody = "abc"
	e.dep.asset.Hash = stringToSHA256(e.responseBody)

	err := e.dep.install()
	assert.NoError(e.T(), err)
}

func (e *DependencyTestSuite) TestInstallBadAssetHash() {
	e.responseBody = "abc"
	e.dep.asset.Hash = "bad bad hash boy"

	err := e.dep.install()
	assert.Error(e.T(), err)
}

func (e *DependencyTestSuite) TestIsCached() {
	cached, err := e.dep.isCached()
	assert.False(e.T(), cached)
	assert.NoError(e.T(), err)

	os.MkdirAll(e.dep.path(), 0755)
	cached, err = e.dep.isCached()
	assert.True(e.T(), cached)
	assert.NoError(e.T(), err)
}

func (e *DependencyTestSuite) TestIsCachedDirIsNotDirectory() {
	os.MkdirAll(path.Dir(e.dep.path()), 0755)
	os.OpenFile(e.dep.path(), os.O_RDONLY|os.O_CREATE, 0666)

	cached, err := e.dep.isCached()
	assert.True(e.T(), cached)
	assert.Error(e.T(), err)
}

func TestInstallDependency(t *testing.T) {
	suite.Run(t, new(DependencyTestSuite))
}

func TestDependencyManager(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}

func stringToSHA256(hash string) string {
	h := sha256.New()
	h.Write([]byte(hash))
	return hex.EncodeToString(h.Sum(nil))
}
