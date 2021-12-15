//go:build integration && !race
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

func TestNamespaceStorage(t *testing.T) {
	testWithEtcd(t, func(s store.Store) {
		ctx := context.Background()

		// We should receive the default namespace (set in store_test.go)
		pred := &store.SelectionPredicate{}
		namespaces, err := s.ListNamespaces(ctx, pred)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(namespaces))

		// We should be able to create a new namespace
		namespace := types.FixtureNamespace("acme")
		err = s.CreateNamespace(ctx, namespace)
		assert.NoError(t, err)

		ctx = context.WithValue(ctx, corev2.NamespaceKey, namespace.Name)

		result, err := s.GetNamespace(ctx, namespace.Name)
		assert.NoError(t, err)
		assert.Equal(t, namespace.Name, result.Name)

		// Missing namespace
		result, err = s.GetNamespace(ctx, "missing")
		assert.NoError(t, err)
		assert.Nil(t, result)

		// Get all namespaces
		namespaces, err = s.ListNamespaces(ctx, pred)
		assert.NoError(t, err)
		assert.NotEmpty(t, namespaces)
		assert.Equal(t, 2, len(namespaces))

		// Delete a non-empty namespace
		check := types.FixtureCheckConfig("entity")
		check.ObjectMeta.Namespace = namespace.Name
		require.NoError(t, s.UpdateCheckConfig(ctx, check))
		err = s.DeleteNamespace(ctx, namespace.Name)
		assert.Error(t, err)
		err = s.DeleteCheckConfigByName(ctx, check.ObjectMeta.Name)
		assert.NoError(t, err)

		// Delete an empty namespace
		err = s.DeleteNamespace(ctx, namespace.Name)
		assert.NoError(t, err)

		// Delete a missing namespace
		err = s.DeleteNamespace(ctx, "missing")
		assert.Error(t, err)

		// Get again all namespaces
		namespaces, err = s.ListNamespaces(ctx, pred)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(namespaces))
	})
}

// TestListNamespacesPagination tests the store's ability to paginate Namespaces.
// While ListNamespaces() internally merely calls the generic List() method of
// the store, we can't rely on that method's tests because they assume a
// generic,"well formed" object with object metadata. Namespace is a snowflake
// in the sense that its object metadata is always empty. So we need to write
// specific tests for it.
func TestListNamespacesPagination(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		for i := 1; i <= 21; i++ {
			// We force the object name to be 2 digits "wide" in order to
			// have a "natural" lexicographic order: 01, 02, ... instead of 1,
			// 11, ...
			objectName := fmt.Sprintf("%.2d", i)
			object := corev2.FixtureNamespace(objectName)

			if err := store.CreateNamespace(context.Background(), object); err != nil {
				t.Fatal(err)
			}
		}

		// We get rid of the default namespace so that it doesn't
		// interfere with our tests below.
		if err := store.DeleteNamespace(context.Background(), "default"); err != nil {
			t.Fatal(err)
		}

		ctx := context.Background()
		t.Run("paginate through namespaces", func(t *testing.T) {
			testListNamespacesPagination(t, ctx, store, 10, 21)
		})

		t.Run("page size equals one", func(t *testing.T) {
			testListNamespacesPagination(t, ctx, store, 1, 21)
		})

		t.Run("page size bigger than set size", func(t *testing.T) {
			testListNamespacesPagination(t, ctx, store, 1337, 21)
		})
	})
}

func testListNamespacesPagination(t *testing.T, ctx context.Context, etcd store.Store, pageSize, setSize int) {
	pred := &store.SelectionPredicate{Limit: int64(pageSize)}
	nFullPages := setSize / pageSize
	nLeftovers := setSize % pageSize

	for i := 0; i < nFullPages; i++ {
		objects, err := etcd.ListNamespaces(ctx, pred)
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
	}

	// Check the last page, supposed to hold nLeftovers objects
	if nLeftovers > 0 {
		objects, err := etcd.ListNamespaces(ctx, pred)
		if err != nil {
			t.Fatal(err)
		}

		if len(objects) != nLeftovers {
			t.Fatalf("Expected last page with %d objects, got %d", nLeftovers, len(objects))
		}

		if pred.Continue != "" {
			t.Fatalf("Expected next continue token to be \"\", got %s", pred.Continue)
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
