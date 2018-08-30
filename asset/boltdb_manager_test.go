package asset_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/types"
)

func TestGetExistingAsset(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "asset_test_get_existing_asset")
	if err != nil {
		log.Printf("unable to create test boltdb file: %v", err)
		t.FailNow()
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

type FailingFetcher struct{}

func (f FailingFetcher) Fetch(path string) (*os.File, error) {
	return nil, fmt.Errorf("failure")
}

func TestGetNonexistentAsset(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "asset_test_get_existing_asset")
	if err != nil {
		log.Printf("unable to create test boltdb file: %v", err)
		t.FailNow()
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
		Fetcher: FailingFetcher{},
	}

	a := &types.Asset{
		URL: "url",
	}

	runtimeAsset, err := manager.Get(a)
	if runtimeAsset != nil {
		t.Failed()
	}

	if err == nil {
		t.Failed()
	}
}

type LocalFetcher struct{}

func (l LocalFetcher) Fetch(path string) (*os.File, error) {
	return os.Open(path)
}

func TestGetInvalidAsset(t *testing.T) {

}

func TestGetInvalidArchive(t *testing.T) {

}

func TestSuccessfulGetExternalAsset(t *testing.T) {

}
