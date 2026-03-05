package asset

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	bolt "go.etcd.io/bbolt"
)

func TestAssetLastAccessedTimestamp(t *testing.T) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "asset_timestamp_test")
	if err != nil {
		t.Fatalf("unable to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a temporary BoltDB database
	db, err := bolt.Open(tmpDir+"/test.db", 0600, nil)
	if err != nil {
		t.Fatalf("unable to create boltdb: %v", err)
	}
	defer db.Close()

	// Create a mock asset manager
	manager := &boltDBAssetManager{
		localStorage: tmpDir,
		db:           db,
		fetcher:      &mockFetcher{pass: true},
		verifier:     &mockVerifier{pass: true},
		expander:     &mockExpander{pass: true},
	}

	// Create a test asset
	asset := &types.Asset{
		URL:    "https://example.com/asset.tar.gz",
		Sha512: "test-sha512",
	}
	asset.Name = "test-asset"

	// First call should create the asset with a timestamp
	startTime := time.Now().Unix()
	runtimeAsset1, err := manager.Get(context.TODO(), asset)
	if err != nil {
		t.Fatalf("unexpected error getting asset: %v", err)
	}

	if runtimeAsset1.LastAccessed < startTime {
		t.Errorf("expected LastAccessed to be recent, got %d, expected >= %d", runtimeAsset1.LastAccessed, startTime)
	}

	// Wait a moment and access again
	time.Sleep(1 * time.Second)
	midTime := time.Now().Unix()

	// Second call should update the timestamp
	runtimeAsset2, err := manager.Get(context.TODO(), asset)
	if err != nil {
		t.Fatalf("unexpected error getting asset: %v", err)
	}

	if runtimeAsset2.LastAccessed < midTime {
		t.Errorf("expected LastAccessed to be updated on second access, got %d, expected >= %d", runtimeAsset2.LastAccessed, midTime)
	}

	if runtimeAsset2.LastAccessed <= runtimeAsset1.LastAccessed {
		t.Errorf("expected second access to have newer timestamp, got %d <= %d", runtimeAsset2.LastAccessed, runtimeAsset1.LastAccessed)
	}
}
