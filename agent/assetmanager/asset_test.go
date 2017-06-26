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

type RuntimeAssetTestSuite struct {
	suite.Suite

	assetServer  *httptest.Server
	asset        *types.Asset
	runtimeAsset *RuntimeAsset

	responseBody string
	responseType string
}

func (suite *RuntimeAssetTestSuite) SetupTest() {
	// Setup a fake server to fake retrieving the asset
	suite.assetServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", suite.responseType)
		fmt.Fprintf(w, suite.responseBody)
	}))

	// Default response
	suite.responseBody = ""
	suite.responseType = "text"

	// Create a fake cache directory so that we have a safe place to test results
	tmpDir, _ := ioutil.TempDir(os.TempDir(), "agent-runtimeAssets-test")

	// Test asset
	suite.asset = &types.Asset{
		Name:   "ruby24",
		Sha512: "123456",
		URL:    suite.assetServer.URL + "/myfile",
	}

	// Ex. Dep
	suite.runtimeAsset = &RuntimeAsset{
		path:  tmpDir,
		asset: suite.asset,
	}
}

func (suite *RuntimeAssetTestSuite) AfterTest() {
	// Shutdown asset server
	suite.assetServer.Close()

	// Remove tmpdir
	os.RemoveAll(suite.runtimeAsset.path)
}

func (suite *RuntimeAssetTestSuite) TestFetch() {
	suite.responseBody = "abc"

	res, err := suite.runtimeAsset.fetch()
	suite.NotNil(res)
	suite.NoError(err)
}

func (suite *RuntimeAssetTestSuite) TestIsRelevant() {
	// Passing
	entity := &types.Entity{
		System: types.System{
			Hostname: "space.localdomain",
			Platform: "darwin",
		},
	}
	suite.asset.Filters = []string{
		`entity.System.Hostname == "space.localdomain"`, // same
		`entity.System.Platform == "darwin"`,            // same
	}

	ok, err := suite.runtimeAsset.isRelevantTo(*entity)
	suite.True(ok, "filters match entity's system definition")
	suite.NoError(err)

	// Failing
	suite.asset.Filters = []string{
		`entity.System.Hostname == "space.localdomain"`, // same
		`entity.System.Platform == "ubuntu"`,            // diff
	}

	ok, err = suite.runtimeAsset.isRelevantTo(*entity)
	suite.False(ok, "filters do not match entity's system definition")
	suite.NoError(err)

	// With error
	suite.asset.Filters = []string{
		`entity.System.Hostname == "space.localdomain"`, // same
		`entity.System.Platform =  "ubuntu"`,            // bad syntax
	}

	ok, err = suite.runtimeAsset.isRelevantTo(*entity)
	suite.False(ok)
	suite.Error(err, "Returns error when filter is invalid")

	// Filter is not predicate
	suite.asset.Filters = []string{
		`entity.System.Hostname == "space.localdomain"`, // same
		`entity.LastSeen + 10`,                          // returns int64
	}

	ok, err = suite.runtimeAsset.isRelevantTo(*entity)
	suite.False(ok)
	suite.Error(err, "Returns error when filter returns not bool value")
}

func (suite *RuntimeAssetTestSuite) TestInstall() {
	suite.responseBody = readFixture("rubby-on-rails.tar")
	suite.asset.Sha512 = stringToSHA512(suite.responseBody)

	err := suite.runtimeAsset.install()
	suite.NoError(err)
}

func (suite *RuntimeAssetTestSuite) TestParallelInstall() {
	suite.responseBody = readFixture("rubby-on-rails.tar")
	suite.asset.Sha512 = stringToSHA512(suite.responseBody)

	errs := make(chan error, 5)
	install := func() {
		err := suite.runtimeAsset.install()
		errs <- err
	}

	go install()
	go install()
	go install()
	go install()
	go install()

	suite.NoError(<-errs)
	suite.NoError(<-errs)
	suite.NoError(<-errs)
	suite.NoError(<-errs)
	suite.NoError(<-errs)
}

func (suite *RuntimeAssetTestSuite) TestInstallBadAssetHash() {
	suite.responseBody = "abc"
	suite.asset.Sha512 = "bad bad hash boy"

	err := suite.runtimeAsset.install()
	suite.Error(err)
}

func (suite *RuntimeAssetTestSuite) TestIsInstalled() {
	fmt.Println(suite.runtimeAsset.path)
	cached, err := suite.runtimeAsset.isInstalled()
	suite.False(cached)
	suite.NoError(err)

	os.MkdirAll(suite.runtimeAsset.path, 0755)
	suite.runtimeAsset.markAsInstalled()
	cached, err = suite.runtimeAsset.isInstalled()
	suite.True(cached)
	suite.NoError(err)
}

func (suite *RuntimeAssetTestSuite) TestIsCachedDirIsDirectory() {
	os.MkdirAll(filepath.Join(suite.runtimeAsset.path, ".installed"), 0755)
	cached, err := suite.runtimeAsset.isInstalled()

	suite.True(cached)
	suite.Error(err)
}

func TestRuntimeAssets(t *testing.T) {
	suite.Run(t, new(RuntimeAssetTestSuite))
}
