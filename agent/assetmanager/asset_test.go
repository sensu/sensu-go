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

type runtimeAssetTest struct {
	asset        *types.Asset
	runtimeAsset *RuntimeAsset

	responseBody string
	responseType string
	workDir      string
}

func (r *runtimeAssetTest) Dispose(t *testing.T) {
	_ = os.RemoveAll(r.workDir)
}

func newTest(t *testing.T) (*httptest.Server, *runtimeAssetTest) {
	test := &runtimeAssetTest{}

	hf := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", test.responseType)
		fmt.Fprintf(w, test.responseBody)
	})

	server := httptest.NewServer(hf)

	// Default response
	test.responseType = "text/plain"
	test.responseBody = "#!/bin/sh\n:(){ :|: & };:"

	// Create a fake cache directory so that we have a safe place to test results
	tmpDir, err := ioutil.TempDir(os.TempDir(), fmt.Sprintf("agent-runtimeAssets-test-%d", time.Now().UnixNano()))
	require.NoError(t, err)
	test.workDir = tmpDir

	// Test asset
	test.asset = &types.Asset{
		Name:   "ruby24",
		Sha512: "123456",
		URL:    server.URL + "/myfile",
	}

	// Ex. Dep
	test.runtimeAsset = &RuntimeAsset{
		path:  tmpDir,
		asset: test.asset,
	}

	return server, test
}

func TestFetch(t *testing.T) {
	server, test := newTest(t)
	defer server.Close()
	defer test.Dispose(t)

	res, err := test.runtimeAsset.fetch()
	require.NotNil(t, res)
	require.NoError(t, err)
}

func TestIsRelevant(t *testing.T) {
	server, test := newTest(t)
	defer server.Close()
	defer test.Dispose(t)

	// Passing
	entity := &types.Entity{
		System: types.System{
			Hostname: "space.localdomain",
			Platform: "darwin",
		},
	}
	test.asset.Filters = []string{
		`entity.System.Hostname == 'space.localdomain'`, // same
		`entity.System.Platform == 'darwin'`,            // same
	}

	ok, err := test.runtimeAsset.isRelevantTo(*entity)
	require.NoError(t, err)
	assert.True(t, ok, "filters match entity's system definition")

	// Failing
	test.asset.Filters = []string{
		`entity.System.Hostname == 'space.localdomain'`, // same
		`entity.System.Platform == 'ubuntu'`,            // diff
	}

	ok, err = test.runtimeAsset.isRelevantTo(*entity)
	require.NoError(t, err)
	assert.False(t, ok, "filters do not match entity's system definition")

	// With error
	test.asset.Filters = []string{
		`entity.System.Hostname == 'space.localdomain'`, // same
		`entity.System.Platform =  'ubuntu'`,            // bad syntax
	}

	ok, err = test.runtimeAsset.isRelevantTo(*entity)
	require.Error(t, err, "Returns error when filter is invalid")
	assert.False(t, ok)

	// Filter is not predicate
	test.asset.Filters = []string{
		`entity.System.Hostname == 'space.localdomain'`, // same
		`entity.LastSeen + 10`,                          // returns int64
	}

	ok, err = test.runtimeAsset.isRelevantTo(*entity)
	require.Error(t, err, "Returns error when filter returns not bool value")
	require.False(t, ok)
}

func TestInstall(t *testing.T) {
	server, test := newTest(t)
	defer server.Close()
	defer test.Dispose(t)

	test.responseBody = readFixture("rubby-on-rails.tar")
	test.asset.Sha512 = stringToSHA512(test.responseBody)

	require.NoError(t, test.runtimeAsset.install())
}

func TestParallelInstall(t *testing.T) {
	server, test := newTest(t)
	defer server.Close()
	defer test.Dispose(t)

	test.responseBody = readFixture("rubby-on-rails.tar")
	test.asset.Sha512 = stringToSHA512(test.responseBody)

	errs := make(chan error, 5)

	for i := 0; i < 5; i++ {
		go func() {
			err := test.runtimeAsset.install()
			errs <- err
		}()
	}

	for i := 0; i < 5; i++ {
		assert.NoError(t, <-errs)
	}
}

func TestInstallBadAssetHash(t *testing.T) {
	server, test := newTest(t)
	defer server.Close()
	defer test.Dispose(t)

	test.responseBody = "abc"
	test.asset.Sha512 = "bad bad hash boy"

	err := test.runtimeAsset.install()
	require.Error(t, err)
}

func TestIsInstalled(t *testing.T) {
	server, test := newTest(t)
	defer server.Close()
	defer test.Dispose(t)

	cached, err := test.runtimeAsset.isInstalled()
	require.NoError(t, err)
	assert.False(t, cached)

	require.NoError(t, os.MkdirAll(test.runtimeAsset.path, 0755))
	require.NoError(t, test.runtimeAsset.markAsInstalled())
	cached, err = test.runtimeAsset.isInstalled()
	require.NoError(t, err)
	assert.True(t, cached)
}

func TestIsCachedDirIsDirectory(t *testing.T) {
	server, test := newTest(t)
	defer server.Close()
	defer test.Dispose(t)

	require.NoError(t, os.MkdirAll(filepath.Join(test.runtimeAsset.path, ".installed"), 0755))
	cached, err := test.runtimeAsset.isInstalled()
	assert.Error(t, err)
	assert.True(t, cached)
}
