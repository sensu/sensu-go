package asset_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/types"
)

func TestGetExistingAsset(t *testing.T) {
	tmpdb, err := ioutil.TempFile(os.TempDir(), "asset_test_get_existing_asset")
	if err != nil {
		log.Printf("unable to create test boltdb file: %v", err)
		t.FailNow()
	}
	defer tmpdb.Close()

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

	db, err := bolt.Open(tmpdb.Name(), 0666, &bolt.Options{})
	if err != nil {
		log.Fatalf("unable to open boltdb in test: %v", err)
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

}

func TestGetInvalidAsset(t *testing.T) {

}

func TestGetInvalidArchive(t *testing.T) {

}

func TestGetExternalAsset(t *testing.T) {

}
