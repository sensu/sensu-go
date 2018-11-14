package asset

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"testing"

	bolt "go.etcd.io/bbolt"

	"github.com/sensu/sensu-go/types"
)

type mockFetcher struct {
	pass bool
}

func (m *mockFetcher) Fetch(string) (*os.File, error) {
	if m.pass {
		return ioutil.TempFile(os.TempDir(), "boltdb_manager_test_fetcher")
	}

	return nil, errors.New("")
}

type mockVerifier struct {
	pass bool
}

func (m *mockVerifier) Verify(f io.ReadSeeker, sha512 string) error {
	if m.pass {
		return nil
	}
	return errors.New("")
}

type mockExpander struct {
	pass bool
}

func (m *mockExpander) Expand(f io.ReadSeeker, path string) error {
	if m.pass {
		return nil
	}

	return errors.New("")
}

func TestGetExistingAsset(t *testing.T) {
	t.Parallel()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "asset_test_get_existing_asset.db")
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

	path := "path"
	sha := "sha"

	a := &types.Asset{
		Sha512: sha,
	}
	runtimeAsset := &RuntimeAsset{
		Path: path,
	}

	runtimeAssetJSON, err := json.Marshal(runtimeAsset)
	if err != nil {
		t.Fatalf("unable to marshal runtime asset in test: %v", err)
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("assets"))
		if err != nil {
			return err
		}
		return bucket.Put([]byte(a.Sha512), runtimeAssetJSON)
	}); err != nil {
		t.Fatalf("unable to update boltdb: %v", err)
	}

	manager := &boltDBAssetManager{
		db: db,
	}

	fetchedAsset, err := manager.Get(a)
	if err != nil {
		t.Logf("expected no error getting asset, got: %v", err)
		t.Failed()
	}

	if fetchedAsset.Path != path {
		t.Logf("expected asset path %s, got %s", runtimeAsset.Path, path)
		t.Failed()
	}
}

func TestGetNonexistentAsset(t *testing.T) {
	t.Parallel()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "asset_test_get_nonexistent_asset.db")
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

	manager := &boltDBAssetManager{
		db:      db,
		fetcher: &mockFetcher{false},
	}

	a := &types.Asset{
		URL: "nonexistent.tar",
	}

	runtimeAsset, err := manager.Get(a)
	if runtimeAsset != nil {
		t.Logf("expected nil runtime asset, got %v", runtimeAsset)
		t.Failed()
	}

	if err == nil {
		t.Log("expected error, got nil")
		t.Failed()
	}
}

func TestGetInvalidAsset(t *testing.T) {
	t.Parallel()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "asset_test_get_invalid_asset.db")
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

	manager := &boltDBAssetManager{
		db:       db,
		fetcher:  &mockFetcher{true},
		verifier: &mockVerifier{false},
	}

	a := &types.Asset{
		URL: "",
	}

	runtimeAsset, err := manager.Get(a)
	if runtimeAsset != nil {
		t.Logf("expected nil runtime asset, got %v", runtimeAsset)
		t.Failed()
	}

	if err == nil {
		t.Log("expected error, got nil")
		t.Failed()
	}
}

func TestFailedExpand(t *testing.T) {
	t.Parallel()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "asset_test_get_invalid_asset.db")
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

	manager := &boltDBAssetManager{
		db:       db,
		fetcher:  &mockFetcher{true},
		verifier: &mockVerifier{true},
		expander: &mockExpander{false},
	}

	a := &types.Asset{
		URL: "",
	}

	runtimeAsset, err := manager.Get(a)
	if runtimeAsset != nil {
		t.Logf("expected nil runtime asset, got %v", runtimeAsset)
		t.Fail()
	}

	if err == nil {
		t.Log("expected error, got nil")
		t.Fail()
	}
}

func TestSuccessfulGetAsset(t *testing.T) {
	t.Parallel()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "asset_test_get_invalid_asset.db")
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

	manager := &boltDBAssetManager{
		db:       db,
		fetcher:  &mockFetcher{true},
		verifier: &mockVerifier{true},
		expander: &mockExpander{true},
	}

	a := &types.Asset{
		ObjectMeta: types.ObjectMeta{
			Name:      "asset",
			Namespace: "default",
		},
		Sha512: "sha",
		URL:    "path",
	}

	runtimeAsset, err := manager.Get(a)
	if err != nil {
		t.Logf("expected no error, got: %v", err)
		t.Fail()
	}

	if runtimeAsset == nil {
		t.Log("expected runtime asset, got nil")
		// can't continue, will panic on runtimeAsset.LibDir()
		t.FailNow()
	}

	d := runtimeAsset.LibDir()

	if d == "" {
		t.Logf("expected lib directory, got: %s", d)
		t.Fail()
	}

	runtimeAsset, err = manager.Get(a)
	if err != nil {
		t.Logf("expected not to receive error, got: %v", err)
		t.Fail()
	}

	if runtimeAsset == nil {
		t.Log("expected runtime asset, got nil")
		// can't continue, will panic on runtimeAsset.LibDir()
		t.FailNow()
	}

	if d2 := runtimeAsset.LibDir(); d2 != d {
		t.Logf("expected lib path to be %s, got %s", d, d2)
		t.Fail()
	}
}
