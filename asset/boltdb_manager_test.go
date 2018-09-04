package asset_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/sensu/sensu-go/asset"
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

func (m *mockVerifier) Verify(f *os.File, sha512 string) error {
	if m.pass {
		return nil
	}
	return errors.New("")
}

type mockExpander struct {
	pass bool
}

func (m *mockExpander) Expand(f *os.File, path string) error {
	if m.pass {
		return nil
	}

	return errors.New("")
}

func TestGetExistingAsset(t *testing.T) {
	t.Parallel()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "asset_test_get_existing_asset.db")
	if err != nil {
		log.Fatalf("unable to create test boltdb file: %v", err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := bolt.Open(tmpFile.Name(), 0666, &bolt.Options{})
	if err != nil {
		log.Fatalf("unable to open boltdb in test: %v", err)
	}
	defer db.Close()

	path := "path"
	sha := "sha"

	a := &types.Asset{
		Sha512: sha,
	}
	runtimeAsset := &asset.RuntimeAsset{
		Path: path,
	}

	runtimeAssetJSON, err := json.Marshal(runtimeAsset)
	if err != nil {
		log.Fatalf("unable to marshal runtime asset in test: %v", err)
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("assets"))
		if err != nil {
			return err
		}
		return bucket.Put([]byte(a.Sha512), runtimeAssetJSON)
	}); err != nil {
		log.Fatalf("unable to update boltdb: %v", err)
	}

	manager := &asset.BoltDBAssetManager{
		DB: db,
	}

	fetchedAsset, err := manager.Get(a)
	if err != nil {
		log.Printf("expected no error getting asset, got: %v", err)
		t.Failed()
	}

	if fetchedAsset.Path != path {
		log.Printf("expected asset path %s, got %s", runtimeAsset.Path, path)
		t.Failed()
	}
}

func TestGetNonexistentAsset(t *testing.T) {
	t.Parallel()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "asset_test_get_nonexistent_asset.db")
	if err != nil {
		log.Fatalf("unable to create test boltdb file: %v", err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := bolt.Open(tmpFile.Name(), 0666, &bolt.Options{})
	if err != nil {
		log.Fatalf("unable to open boltdb in test: %v", err)
	}
	defer db.Close()

	manager := &asset.BoltDBAssetManager{
		DB:      db,
		Fetcher: &mockFetcher{false},
	}

	a := &types.Asset{
		URL: "nonexistent.tar",
	}

	runtimeAsset, err := manager.Get(a)
	if runtimeAsset != nil {
		t.Failed()
	}

	if err == nil {
		t.Failed()
	}
}

func TestGetInvalidAsset(t *testing.T) {
	t.Parallel()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "asset_test_get_invalid_asset.db")
	if err != nil {
		log.Fatalf("unable to create test boltdb file: %v", err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := bolt.Open(tmpFile.Name(), 0666, &bolt.Options{})
	if err != nil {
		log.Fatalf("unable to open boltdb in test: %v", err)
	}
	defer db.Close()

	manager := &asset.BoltDBAssetManager{
		DB:       db,
		Fetcher:  &mockFetcher{true},
		Verifier: &mockVerifier{false},
	}

	a := &types.Asset{
		URL: "",
	}

	runtimeAsset, err := manager.Get(a)
	if runtimeAsset != nil {
		t.Failed()
	}

	if err == nil {
		t.Failed()
	}
}

func TestFailedExpand(t *testing.T) {
	t.Parallel()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "asset_test_get_invalid_asset.db")
	if err != nil {
		log.Fatalf("unable to create test boltdb file: %v", err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := bolt.Open(tmpFile.Name(), 0666, &bolt.Options{})
	if err != nil {
		log.Fatalf("unable to open boltdb in test: %v", err)
	}
	defer db.Close()

	manager := &asset.BoltDBAssetManager{
		DB:       db,
		Fetcher:  &mockFetcher{true},
		Verifier: &mockVerifier{true},
		Expander: &mockExpander{false},
	}

	a := &types.Asset{
		URL: "",
	}

	runtimeAsset, err := manager.Get(a)
	if runtimeAsset != nil {
		log.Println("received asset, expected nil")
		t.FailNow()
	}

	if err == nil {
		log.Println("received nil error, expected error")
	}
}

func TestSuccessfulGetAsset(t *testing.T) {
	t.Parallel()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "asset_test_get_invalid_asset.db")
	if err != nil {
		log.Fatalf("unable to create test boltdb file: %v", err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := bolt.Open(tmpFile.Name(), 0666, &bolt.Options{})
	if err != nil {
		log.Fatalf("unable to open boltdb in test: %v", err)
	}
	defer db.Close()

	manager := &asset.BoltDBAssetManager{
		DB:       db,
		Fetcher:  &mockFetcher{true},
		Verifier: &mockVerifier{true},
		Expander: &mockExpander{true},
	}

	a := &types.Asset{
		Name:   "asset",
		Sha512: "sha",
		URL:    "path",
	}

	runtimeAsset, err := manager.Get(a)
	if runtimeAsset == nil {
		log.Println("expected asset, got nil")
		t.FailNow()
	}

	if err != nil {
		log.Printf("expected no error, got: %v", err)
		t.FailNow()
	}

	d := runtimeAsset.LibDir()

	if d == "" {
		log.Printf("expected lib directory, got: %s", d)
		t.FailNow()
	}

	runtimeAsset, err = manager.Get(a)
	if err != nil {
		log.Printf("expected not to receive error, got: %v", err)
		t.FailNow()
	}

	if d2 := runtimeAsset.LibDir(); d2 != d {
		log.Printf("expected lib path to be %s, got %s", d, d2)
		t.FailNow()
	}
}
