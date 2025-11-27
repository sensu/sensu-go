package asset

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/prometheus/client_golang/prometheus"
	corev2 "github.com/sensu/core/v2"
	metricspkg "github.com/sensu/sensu-go/metrics"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/time/rate"
)

const (
	// FetchDuration is the name of the prometheus summary vec used to track
	// average latencies of asset fetching.
	FetchDuration = "sensu_go_asset_fetch_duration"

	// ExpandDuration is the name of the prometheus summary vec used to track
	// average latencies of asset expansion.
	ExpandDuration = "sensu_go_asset_expand_duration"
)

var (
	assetBucketName = []byte("assets")

	fetchDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       FetchDuration,
			Help:       "asset fetching latency distribution",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{metricspkg.StatusLabelName, "name", "namespace"},
	)

	expandDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       ExpandDuration,
			Help:       "asset expansion latency distribution",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{metricspkg.StatusLabelName, "name", "namespace"},
	)
)

func init() {
	if err := prometheus.Register(fetchDuration); err != nil {
		panic(metricspkg.FormatRegistrationErr(FetchDuration, err))
	}
	if err := prometheus.Register(expandDuration); err != nil {
		panic(metricspkg.FormatRegistrationErr(ExpandDuration, err))
	}
}

func GetBoltDBConnection(cacheDir string) (*bolt.DB, error) {
	// create agent cache directory if it doesn't already exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	logger.WithField("cache", cacheDir).Debug("initializing cache directory")
	db, err := bolt.Open(filepath.Join(cacheDir, dbName), 0600, &bolt.Options{Timeout: 60 * time.Second})
	if err != nil {
		return nil, err
	}
	logger.WithField("cache", cacheDir).Debug("done initializing cache directory")
	return db, nil
}

// NewBoltDBGetter returns a new default asset Getter. If fetcher, verifier, or
// expander are nil, the getter will use the built-in components.
func NewBoltDBGetter(db *bolt.DB,
	localStorage string,
	trustedCAFile string,
	fetcher Fetcher,
	verifier Verifier,
	expander Expander,
	limiter *rate.Limiter) Getter {

	if fetcher == nil {
		fetcher = &httpFetcher{
			Limiter:       limiter,
			trustedCAFile: trustedCAFile,
		}
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
	}
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
func (b *boltDBAssetManager) Get(ctx context.Context, asset *corev2.Asset) (*RuntimeAsset, error) {
	key := []byte(asset.GetSha512())
	var localAsset *RuntimeAsset

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
		localAsset.Name = asset.Name
		localAsset.SHA512 = asset.Sha512
		// Update last accessed timestamp and persist to database
		if err := b.updateLastAccessed(key, localAsset); err != nil {
			// Log the error but don't fail the asset retrieval
			logger.WithError(err).Debug("failed to update asset last accessed timestamp")
		}
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
				// Update last accessed timestamp for this asset
				localAsset.LastAccessed = time.Now().Unix()

				// Re-serialize and store the updated asset
				if updatedJSON, marshalErr := json.Marshal(localAsset); marshalErr == nil {
					bucket.Put(key, updatedJSON)
				}
				return nil
			}
		}

		// install the asset
		tmpFile, err := b.fetchWithDuration(ctx, asset)
		if err != nil {
			return err
		}
		defer tmpFile.Close()
		defer os.Remove(tmpFile.Name())

		// verify
		if err := b.verifier.Verify(tmpFile, asset.Sha512); err != nil {
			// Attempt to retrieve the size of the downloaded asset
			var size uint64
			if fileInfo, err := tmpFile.Stat(); err == nil {
				size = uint64(fileInfo.Size())
			}

			return fmt.Errorf(
				"could not validate downloaded asset %q (%s): %s",
				asset.Name, humanize.Bytes(size), err,
			)
		}

		// expand
		assetPath, err := b.expandWithDuration(tmpFile, asset)
		if err != nil {
			return err
		}

		localAsset = &RuntimeAsset{
			Path:         assetPath,
			LastAccessed: time.Now().Unix(),
		}

		assetJSON, err := json.Marshal(localAsset)
		if err != nil {
			panic(err)
		}

		return bucket.Put(key, assetJSON)
	}); err != nil {
		return nil, err
	}

	if localAsset != nil {
		localAsset.Name = asset.Name
		localAsset.SHA512 = asset.Sha512
	}

	return localAsset, nil
}

func (b *boltDBAssetManager) fetchWithDuration(ctx context.Context, asset *corev2.Asset) (file *os.File, err error) {
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		status := metricspkg.StatusLabelSuccess
		if err != nil {
			status = metricspkg.StatusLabelError
		}
		fetchDuration.
			WithLabelValues(status, asset.ObjectMeta.Name, asset.ObjectMeta.Namespace).
			Observe(v * float64(1000))
	}))
	defer timer.ObserveDuration()

	return b.fetcher.Fetch(ctx, asset.URL, asset.Headers)
}

func (b *boltDBAssetManager) expandWithDuration(tmpFile *os.File, asset *corev2.Asset) (assetPath string, err error) {
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		status := metricspkg.StatusLabelSuccess
		if err != nil {
			status = metricspkg.StatusLabelError
		}
		expandDuration.
			WithLabelValues(status, asset.ObjectMeta.Name, asset.ObjectMeta.Namespace).
			Observe(v * float64(1000))
	}))
	defer timer.ObserveDuration()

	assetPath = filepath.Join(b.localStorage, asset.Sha512)
	return assetPath, b.expander.Expand(tmpFile, assetPath)
}

// updateLastAccessed updates the LastAccessed timestamp for an asset in the database
func (b *boltDBAssetManager) updateLastAccessed(key []byte, runtimeAsset *RuntimeAsset) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(assetBucketName)
		if bucket == nil {
			return fmt.Errorf("asset bucket not found")
		}

		// Update the timestamp
		runtimeAsset.LastAccessed = time.Now().Unix()

		// Serialize and store the updated asset
		assetJSON, err := json.Marshal(runtimeAsset)
		if err != nil {
			return fmt.Errorf("failed to marshal asset: %w", err)
		}

		return bucket.Put(key, assetJSON)
	})
}

func (b *boltDBAssetManager) GetDB() *bolt.DB {
	return b.db
}

// FindUnusedAssets scans the database for assets older than cutoffTime
func FindUnusedAssets(db *bolt.DB, cutoffTime int64) ([]RuntimeAsset, error) {
	var unusedAssets []RuntimeAsset

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(assetBucketName)
		if bucket == nil {
			logger.Error("no assets bucket found in database")
			return nil
		}

		return bucket.ForEach(func(key, value []byte) error {
			var runtimeAsset RuntimeAsset

			// Unmarshal the asset data
			if err := json.Unmarshal(value, &runtimeAsset); err != nil {
				// Log and skip corrupted entries
				logger.WithError(err).WithField("sha512", string(key)).Warn("skipping corrupted asset entry")
				return nil
			}

			// Check if asset is expired
			if runtimeAsset.LastAccessed != 0 && runtimeAsset.LastAccessed < cutoffTime {
				unusedAssets = append(unusedAssets, runtimeAsset)
			}

			return nil
		})
	})

	return unusedAssets, err
}

// DeleteAsset deletes from database and cache
func DeleteAsset(db *bolt.DB, runtimeAsset RuntimeAsset) error {
	key := []byte(runtimeAsset.SHA512)

	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(assetBucketName)
		if bucket == nil {
			// Nothing to delete
			return nil
		}

		value := bucket.Get(key)
		if value == nil {
			// No matching asset
			return nil
		}

		// Delete the record
		if err := bucket.Delete(key); err != nil {
			return fmt.Errorf("failed to delete asset with sha512 %s: %w", runtimeAsset.SHA512, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to remove asset from database: %w", err)
	}

	// Remove from filesystem
	if runtimeAsset.Path != "" {
		if err := os.RemoveAll(runtimeAsset.Path); err != nil {
			// Log the filesystem error but don't fail the operation
			// since the database entry is already removed
			logger.WithError(err).WithField("path", runtimeAsset.Path).Warn("failed to remove asset from filesystem during cleanup")
		}
	}

	return nil
}
