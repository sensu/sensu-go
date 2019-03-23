// +build integration,!race

package etcd

import (
	"context"
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestExtensionStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		ext := types.FixtureExtension("frobber")
		ctx := context.WithValue(context.Background(), types.NamespaceKey, ext.Namespace)

		err := store.RegisterExtension(ctx, ext)
		assert.NoError(t, err)

		retrieved, err := store.GetExtension(ctx, "frobber")
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, ext.Name, retrieved.Name)
		assert.Equal(t, ext.URL, retrieved.URL)

		extensions, continueToken, err := store.GetExtensions(ctx, 0, "")
		require.NoError(t, err)
		assert.NotEmpty(t, extensions)
		assert.Equal(t, 1, len(extensions))
		assert.Empty(t, continueToken)

		// Updating an ext in a nonexistent org should not work
		ext.Namespace = "missing"
		err = store.RegisterExtension(ctx, ext)
		assert.Error(t, err)
	})
}

func TestGetExtensionsPagination(t *testing.T) {
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
			object := corev2.FixtureExtension(objectName)

			ctx := context.WithValue(context.Background(), types.NamespaceKey, object.Namespace)
			if err := store.RegisterExtension(ctx, object); err != nil {
				t.Fatal(err)
			}

			object.Namespace = "testing"
			ctx = context.WithValue(context.Background(), types.NamespaceKey, object.Namespace)
			if err := store.RegisterExtension(ctx, object); err != nil {
				t.Fatal(err)
			}
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")
		t.Run("objects in default namespace", func(t *testing.T) {
			testGetExtensionsPagination(t, ctx, store, 10, 21)
		})

		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "testing")
		t.Run("objects in testing namespace", func(t *testing.T) {
			testGetExtensionsPagination(t, ctx, store, 10, 21)
		})

		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")
		t.Run("page size equals one", func(t *testing.T) {
			testGetExtensionsPagination(t, ctx, store, 1, 21)
		})

		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")
		t.Run("page size bigger than set size", func(t *testing.T) {
			testGetExtensionsPagination(t, ctx, store, 1337, 21)
		})
	})
}

func testGetExtensionsPagination(t *testing.T, ctx context.Context, etcd store.Store, pageSize, setSize int) {
	nFullPages := setSize / pageSize
	nLeftovers := setSize % pageSize

	continueToken := ""
	for i := 0; i < nFullPages; i++ {
		objects, nextContinueToken, err := etcd.GetExtensions(ctx, int64(pageSize), continueToken)
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
		objects, nextContinueToken, err := etcd.GetExtensions(ctx, int64(pageSize), continueToken)
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
