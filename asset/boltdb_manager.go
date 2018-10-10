package asset

import (
	"encoding/json"
	"os"
	"path/filepath"

	bolt "go.etcd.io/bbolt"

	"github.com/sensu/sensu-go/types"
)

var (
	assetBucketName = []byte("assets")
)

// A Getter is responsible for fetching (based on fitler selection), verifying,
// and expanding an asset. Calls to the Get method block until the Asset has
// fetched, verified, and expanded or it returns an error indicating why getting
// the asset failed.
type Getter interface {
	Get(*types.Asset) (*RuntimeAsset, error)
}

// NewBoltDBGetter returns a new default asset Getter. If fetcher, verifier, or
// expander are nil, the getter will use the built-in components.
func NewBoltDBGetter(db *bolt.DB,
	localStorage string,
	fetcher Fetcher,
	verifier Verifier,
	expander Expander) (Getter, error) {

	if fetcher == nil {
		fetcher = defaultFetcher
	}

	if expander == nil {
		expander = defaultExpander
	}

	if verifier == nil {
		verifier = defaultVerifier
	}

	return &boltDBAssetManager{
		localStorage: localStorage,
		db:           db,
		fetcher:      fetcher,
		expander:     expander,
		verifier:     verifier,
	}, nil
}

// boltDBAssetManager is responsible for the installing and storing the metadata
// for assets backed by an instance of BoltDB on the local filesystem. BoltDB
// provides the serialization guarantee that the asset contract specifies.
// We rely on long-lived BoltDB transactions during Get to provide this
// mechanism for blocking.
type boltDBAssetManager struct {
	localStorage string
	db           *bolt.DB
	fetcher      Fetcher
	expander     Expander
	verifier     Verifier
}

// Get opens a transaction to BoltDB, causing subsequent calls to
// Get to block. During this transaction, we attempt to determine if the asset
// is installed by querying BoltDB for the asset's SHA (which we use as an ID).
//
// If a value is returned, we return the deserialized asset stored in BoltDB.
// If deserialization fails, we assume there is some level of corruption and
// attempt to re-install the asset.
//
// If a value is not returned, the asset is not installed or not installed
// correctly. We then proceed to attempt asset installation.
func (b *boltDBAssetManager) Get(asset *types.Asset) (*RuntimeAsset, error) {
	var localAsset *RuntimeAsset
	key := []byte(asset.Sha512)

	// Concurrent calls to View are allowed, but a concurrent call that has
	// has proceeded to Update below will block here.
	if err := b.db.View(func(tx *bolt.Tx) error {
		// If the key exists, the bucket should already exist.
		bucket := tx.Bucket(assetBucketName)
		if bucket == nil {
			return nil
		}

		value := bucket.Get(key)
		if value != nil {
			// deserialize asset
			if err := json.Unmarshal(value, &localAsset); err == nil {
				return nil
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// Check to see if the view was successful.
	if localAsset != nil {
		return localAsset, nil
	}

	if err := b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(assetBucketName)
		if err != nil {
			return err
		}

		// Though we've already attempted to do this, it's possible that a previous
		// call completed installation of the asset while this transaction
		// was blocked on serialization. Re-attempt to get the key in case that is
		// what happened.
		value := bucket.Get(key)
		if value != nil {
			// deserialize asset
			if err := json.Unmarshal(value, &localAsset); err == nil {
				return nil
			}
		}

		// install the asset
		tmpFile, err := b.fetcher.Fetch(asset.URL)
		if err != nil {
			return err
		}
		defer tmpFile.Close()
		defer os.Remove(tmpFile.Name())

		// verify
		if err := b.verifier.Verify(tmpFile, asset.Sha512); err != nil {
			return err
		}

		// expand
		assetPath := filepath.Join(b.localStorage, asset.Sha512)
		if err := b.expander.Expand(tmpFile, assetPath); err != nil {
			return err
		}

		localAsset = &RuntimeAsset{
			Path: assetPath,
		}

		assetJSON, err := json.Marshal(localAsset)
		if err != nil {
			panic(err)
		}

		return bucket.Put(key, assetJSON)
	}); err != nil {
		return nil, err
	}

	return localAsset, nil
}
