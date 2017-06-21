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
	dep          *RuntimeAsset
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
	tmpDir, _ := ioutil.TempDir(os.TempDir(), "agent-deps-test")

	// Ex. Dep
	suite.dep = &RuntimeAsset{
		path: tmpDir,
		asset: &types.Asset{
			Name:   "ruby24",
			Sha512: "123456",
			URL:    suite.assetServer.URL + "/myfile",
		},
	}
}

func (suite *RuntimeAssetTestSuite) AfterTest() {
	// Shutdown asset server
	suite.assetServer.Close()

	// Remove tmpdir
	os.RemoveAll(suite.dep.path)
}

func (suite *RuntimeAssetTestSuite) TestFetch() {
	suite.responseBody = "abc"

	res, err := suite.dep.fetch()
	suite.NotNil(res)
	suite.NoError(err)
}

func (suite *RuntimeAssetTestSuite) TestInstall() {
	suite.responseBody = readFixture("rubby-on-rails.tar")
	suite.dep.asset.Sha512 = stringToSHA512(suite.responseBody)

	err := suite.dep.install()
	suite.NoError(err)
}

func (suite *RuntimeAssetTestSuite) TestParallelInstall() {
	suite.responseBody = readFixture("rubby-on-rails.tar")
	suite.dep.asset.Sha512 = stringToSHA512(suite.responseBody)

	errs := make(chan error, 5)
	install := func() {
		err := suite.dep.install()
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
	suite.dep.asset.Sha512 = "bad bad hash boy"

	err := suite.dep.install()
	suite.Error(err)
}

func (suite *RuntimeAssetTestSuite) TestIsInstalled() {
	fmt.Println(suite.dep.path)
	cached, err := suite.dep.isInstalled()
	suite.False(cached)
	suite.NoError(err)

	os.MkdirAll(suite.dep.path, 0755)
	suite.dep.markAsInstalled()
	cached, err = suite.dep.isInstalled()
	suite.True(cached)
	suite.NoError(err)
}

func (suite *RuntimeAssetTestSuite) TestIsCachedDirIsDirectory() {
	os.MkdirAll(filepath.Join(suite.dep.path, ".installed"), 0755)
	cached, err := suite.dep.isInstalled()

	suite.True(cached)
	suite.Error(err)
}

func TestRuntimeAssets(t *testing.T) {
	suite.Run(t, new(RuntimeAssetTestSuite))
}
