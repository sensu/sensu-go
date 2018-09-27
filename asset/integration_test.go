package asset_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	bolt "go.etcd.io/bbolt"

	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/types"
)

type localFetcher struct{}

func (f *localFetcher) Fetch(path string) (*os.File, error) {
	return os.Open(path)
}

func TestBoltDBManager(t *testing.T) {
	t.Parallel()

	tmpDir := os.TempDir()

	path, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("error getting cwd: %v", err)
	}
	fixturePath := filepath.Join(path, "fixtures", "rubby-on-rails.tgz")
	tmpFixturePath := filepath.Join(tmpDir, "rubby-on-rails.tgz")
	// the asset manager actually deletes the "downloaded" archive
	// for us, because it's a temporary file. So create a link to the
	// original asset in the temp dir and use that as our asset "URL".
	srcBytes, err := ioutil.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("unable to read fixture: %v", err)
	}
	if err := ioutil.WriteFile(tmpFixturePath, srcBytes, 0644); err != nil {
		t.Fatalf("unable to write tmp fixture: %v", err)
	}

	fixtureSHAPath := fmt.Sprintf("%s.sha512", fixturePath)
	sha512, err := ioutil.ReadFile(fixtureSHAPath)
	if err != nil {
		t.Fatalf("error reading fixture sha: %v", err)
	}

	fixtureAsset := &types.Asset{
		Sha512: string(sha512),
		URL:    tmpFixturePath,
		Name:   "rubby-on-rails",
	}

	tmpFile, err := ioutil.TempFile(tmpDir, "asset_integration_test.db")
	if err != nil {
		t.Fatalf("unable to create test boltdb file: %v", err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := bolt.Open(tmpFile.Name(), 0666, &bolt.Options{})
	if err != nil {
		t.Fatalf("unable to open boltdb in test: %v", err)
	}
	defer db.Close()

	tmpAssetDir := filepath.Join(tmpDir, "asset_integration_test")
	if err := os.Mkdir(tmpAssetDir, 0755); err != nil {
		t.Fatalf("unable to create directory: %v", err)
	}
	defer os.RemoveAll(tmpAssetDir)

	getter, err := asset.NewBoltDBGetter(
		db,
		tmpAssetDir,
		&localFetcher{},
		nil,
		nil,
	)
	if err != nil {
		t.Logf("error returned, expected nil: %v", err)
		t.Fail()
	}

	if getter == nil {
		t.Log("expected getter to be returned, got nil")
		t.Fail()
	}

	runtimeAsset, err := getter.Get(fixtureAsset)
	if err != nil {
		t.Logf("error getting runtime asset, expected nil: %v", err)
		t.Fail()
	}
	if runtimeAsset == nil {
		t.Log("expected runtime asset, got nil")
		t.FailNow()
	}

	railsPath := filepath.Join(runtimeAsset.BinDir(), "rails")
	railsFile, err := os.Open(railsPath)
	if err != nil {
		t.Logf("expected no error opening rails file, got error: %v", err)
		t.Fail()
	}
	if railsFile == nil {
		t.Log("expected opened rails file, got nil")
		t.FailNow()
	}
	defer railsFile.Close()
}
