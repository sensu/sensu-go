// +build integration,!race

package etcd

import (
	"context"
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestAssetStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		asset := types.FixtureAsset("ruby")
		ctx := context.WithValue(context.Background(), types.NamespaceKey, asset.Namespace)

		err := store.UpdateAsset(ctx, asset)
		assert.NoError(t, err)

		retrieved, err := store.GetAssetByName(ctx, "ruby")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)

		assert.Equal(t, asset.Name, retrieved.Name)
		assert.Equal(t, asset.URL, retrieved.URL)
		assert.Equal(t, asset.Sha512, retrieved.Sha512)

		assets, continueToken, err := store.GetAssets(ctx, 0, "")
		assert.NoError(t, err)
		assert.NotEmpty(t, assets)
		assert.Equal(t, 1, len(assets))
		assert.Empty(t, continueToken)

		// Updating an asset in a nonexistent org should not work
		asset.Namespace = "missing"
		err = store.UpdateAsset(ctx, asset)
		assert.Error(t, err)
	})
}

func TestGetAssetsPagination(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		// Create a "testing" namespace in the store
		testingNS := corev2.FixtureNamespace("testing")
		store.UpdateNamespace(context.Background(), testingNS)

		// Add 42 objects in the store: 21 in the "default" namespace and 21 in
		// the "testing" namespace
		for i := 1; i <= 21; i++ {
			// We force the object name to be 2 digits "wide" in order to
			// have a "natural" lexicographic order: 01, 02, ... instead of 1,
			// 11, ...
			objectName := fmt.Sprintf("%.2d", i)
			object := corev2.FixtureAsset(objectName)

			if err := store.UpdateAsset(context.Background(), object); err != nil {
				t.Fatal(err)
			}

			object.Namespace = "testing"
			if err := store.UpdateAsset(context.Background(), object); err != nil {
				t.Fatal(err)
			}
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")
		t.Run("objects in default namespace", func(t *testing.T) {
			testGetAssetsPagination(t, ctx, store, 10, 21)
		})

		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "testing")
		t.Run("objects in testing namespace", func(t *testing.T) {
			testGetAssetsPagination(t, ctx, store, 10, 21)
		})

		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")
		t.Run("page size equals one", func(t *testing.T) {
			testGetAssetsPagination(t, ctx, store, 1, 21)
		})

		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")
		t.Run("page size bigger than set size", func(t *testing.T) {
			testGetAssetsPagination(t, ctx, store, 1337, 21)
		})
	})
}

func testGetAssetsPagination(t *testing.T, ctx context.Context, etcd store.Store, pageSize, setSize int) {
	nFullPages := setSize / pageSize
	nLeftovers := setSize % pageSize

	continueToken := ""
	for i := 0; i < nFullPages; i++ {
		objects, nextContinueToken, err := etcd.GetAssets(ctx, int64(pageSize), continueToken)
		if err != nil {
			t.Fatal(err)
		}

		if len(objects) != pageSize {
			t.Fatalf("Expected page %d to have %d objects but got %d", i, pageSize, len(objects))
		}

		offset := i * pageSize
		for j, object := range objects {
			n := ((offset + j) % setSize) + 1
			expected := fmt.Sprintf("%.2d", n)

			if object.Name != expected {
				t.Fatalf("Expected %s, got %s", expected, object.Name)
			}
		}

		continueToken = nextContinueToken
	}

	// Check the last page, supposed to hold nLeftovers objects
	if nLeftovers > 0 {
		objects, nextContinueToken, err := etcd.GetAssets(ctx, int64(pageSize), continueToken)
		if err != nil {
			t.Fatal(err)
		}

		if len(objects) != nLeftovers {
			t.Fatalf("Expected last page with %d objects, got %d", nLeftovers, len(objects))
		}

		if nextContinueToken != "" {
			t.Fatalf("Expected next continue token to be \"\", got %s", nextContinueToken)
		}

		offset := pageSize * nFullPages
		for j, object := range objects {
			n := ((offset + j) % setSize) + 1
			expected := fmt.Sprintf("%.2d", n)

			if object.Name != expected {
				t.Fatalf("Expected %s, got %s", expected, object.Name)
			}
		}
	}
}
