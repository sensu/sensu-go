package asset

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"

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
	FlagCacheDir   = "cache-dir"
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
				logger.Println(err)
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

	assetSHA := asset.Sha512
	CacheDir := viper.GetString(FlagCacheDir)
	fullPath := filepath.Join(CacheDir, assetSHA)

	if err := CleanUp(fullPath); err != nil { //fix for git issue 5009
		logger.Println("error cleaning up the SHA dir: %s", err)
	}

	assetPath = filepath.Join(b.localStorage, asset.Sha512)
	return assetPath, b.expander.Expand(tmpFile, assetPath)
}
