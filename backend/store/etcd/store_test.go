//go:build integration && !race
// +build integration,!race

package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd/kvc"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/client/v3"
)

func testWithEtcd(t *testing.T, f func(store.Store)) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()

	s := NewStore(client, e.Name())

	// Mock a default namespace
	require.NoError(t, s.CreateNamespace(context.Background(), types.FixtureNamespace("default")))

	f(s)
}

func testWithEtcdStore(t *testing.T, f func(*Store)) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()

	s := NewStore(client, e.Name())

	// Mock a default namespace
	require.NoError(t, s.CreateNamespace(context.Background(), types.FixtureNamespace("default")))

	f(s)
}

func testWithEtcdClient(t *testing.T, f func(store.Store, *clientv3.Client)) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()

	s := NewStore(client, e.Name())

	// Mock a default namespace
	require.NoError(t, s.CreateNamespace(context.Background(), types.FixtureNamespace("default")))

	f(s, client)
}

func TestCreate(t *testing.T) {
	testWithEtcdStore(t, func(s *Store) {
		// Creating a namespaced key that does not exist should work
		obj := &GenericObject{}
		ctx := context.WithValue(context.Background(), types.NamespaceKey, "default")
		err := Create(ctx, s.client, "/default/foo", "default", obj)
		assert.NoError(t, err)

		// Creating a wrapped resource should work
		err = Create(ctx, s.client, "/default/bar", "default", types.Wrapper{Value: obj})
		assert.NoError(t, err)

		// Creating this same key should return an error that it already exist
		err = Create(ctx, s.client, "/default/foo", "default", obj)
		switch err := err.(type) {
		case *store.ErrAlreadyExists:
			break
		default:
			t.Errorf("Expected error ErrAlreadyExists, received %v", err)
		}

		// Creating a namespaced key in a missing namespace should return an error
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "acme")
		err = Create(ctx, s.client, "/acme/foo", "acme", obj)
		switch err := err.(type) {
		case *store.ErrNamespaceMissing:
			break
		default:
			t.Errorf("Expected error ErrNamespaceMissing, received %v", err)
		}

		// We should also be able to create a global object
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "")
		err = Create(ctx, s.client, "/foo", "", obj)
		assert.NoError(t, err)

		// Creating this same key should return an error that it already exists, and
		// not that the namespace is missing
		err = Create(ctx, s.client, "/foo", "", obj)
		switch err := err.(type) {
		case *store.ErrAlreadyExists:
			break
		default:
			t.Errorf("Expected error ErrAlreadyExists, received %v", err)
		}
	})
}

func TestCreateOrUpdate(t *testing.T) {
	testWithEtcdStore(t, func(s *Store) {
		// Creating a namespaced key that does not exist should work
		obj := &GenericObject{Revision: 1}
		ctx := context.WithValue(context.Background(), types.NamespaceKey, "default")
		err := CreateOrUpdate(ctx, s.client, "/default/foo", "default", obj)
		assert.NoError(t, err)

		// Creating a namespaced key in a missing namespace should return an error
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "acme")
		err = CreateOrUpdate(ctx, s.client, "/acme/foo", "acme", obj)
		switch err := err.(type) {
		case *store.ErrNamespaceMissing:
			break
		default:
			t.Errorf("Expected error ErrNamespaceMissing, received %v", err)
		}

		// Creating this same key should also work, but the revision should be
		// different
		obj2 := &GenericObject{Revision: 2}
		err = CreateOrUpdate(ctx, s.client, "/default/foo", "default", obj)
		assert.NoError(t, err)
		result := &GenericObject{}
		err = Get(ctx, s.client, "/default/foo", obj2)
		assert.NoError(t, err)
		assert.NotEqual(t, obj.Revision, result.Revision)

		// We should also be able to create a global object
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "")
		err = CreateOrUpdate(ctx, s.client, "/foo", "", obj)
		assert.NoError(t, err)
	})
}

func TestDelete(t *testing.T) {
	testWithEtcdStore(t, func(store *Store) {
		// Deleting a non-existant key should fail
		ctx := context.WithValue(context.Background(), types.NamespaceKey, "default")
		require.Error(t, Delete(ctx, store.client, "/default/foo"))

		// Create it first
		obj := &GenericObject{}
		require.NoError(t, Create(ctx, store.client, "/default/foo", "default", obj))

		// Now make sure it gets properly deleted
		require.NoError(t, Delete(ctx, store.client, "/default/foo"))
		result := &GenericObject{}
		require.Error(t, Get(ctx, store.client, "/default/foo", result))
	})
}
func TestGet(t *testing.T) {
	testWithEtcdStore(t, func(store *Store) {
		// Create a namespaced key
		obj := &GenericObject{Revision: 1}
		ctx := context.WithValue(context.Background(), types.NamespaceKey, "default")
		err := Create(ctx, store.client, "/default/foo", "default", obj)
		assert.NoError(t, err)

		// Retrieve the namespaced key and make sure it's the expected object
		result := &GenericObject{}
		err = Get(ctx, store.client, "/default/foo", result)
		assert.NoError(t, err)
		assert.Equal(t, obj, result)

		// Create a global key
		obj2 := &GenericObject{Revision: 2}
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "")
		err = Create(ctx, store.client, "/foo", "", obj2)
		assert.NoError(t, err)

		// Retrieve the global key and make sure it's the expected object
		result2 := &GenericObject{}
		err = Get(ctx, store.client, "/foo", result2)
		assert.NoError(t, err)
		assert.Equal(t, obj2, result2)
	})
}

var genericKeyBuilder = store.NewKeyBuilder("generic")

func getGenericObjectPath(obj *GenericObject) string {
	return genericKeyBuilder.WithResource(obj).Build(obj.Name)
}

func getGenericObjectsPath(ctx context.Context, name string) string {
	return genericKeyBuilder.WithContext(ctx).Build(name)
}
func TestList(t *testing.T) {
	testWithEtcdStore(t, func(s *Store) {
		// Create new namespaces
		require.NoError(t, s.CreateNamespace(context.Background(), types.FixtureNamespace("acme")))
		require.NoError(t, s.CreateNamespace(context.Background(), types.FixtureNamespace("acme-devel")))

		// Create a bunch of keys everywhere
		obj1 := &GenericObject{ObjectMeta: corev2.ObjectMeta{Name: "obj1", Namespace: "default"}}
		ctx := context.WithValue(context.Background(), types.NamespaceKey, "default")
		require.NoError(t, Create(ctx, s.client, "/sensu.io/generic/default/obj1", "default", obj1))

		// Let's encode a resources with JSON (instead of Protobuf) to ensure the
		// List function supports both encoding format
		obj2 := &GenericObject{ObjectMeta: corev2.ObjectMeta{Name: "obj2", Namespace: "acme"}}
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "acme")
		bytes, _ := json.Marshal(obj2)
		_, err := s.client.Put(ctx, "/sensu.io/generic/acme/obj2", string(bytes))
		require.NoError(t, err)

		obj3 := &GenericObject{ObjectMeta: corev2.ObjectMeta{Name: "obj3", Namespace: "acme"}}
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "acme")
		require.NoError(t, Create(ctx, s.client, "/sensu.io/generic/acme/obj3", "acme", obj3))

		// This object is required to test
		// https://github.com/sensu/sensu-enterprise-go/issues/418. We want to make
		// sure resources within a namespace, whose name contains an another
		// namespace as a prefix, are not showing up (e.g. acme & acme-devel)
		obj4 := &GenericObject{ObjectMeta: corev2.ObjectMeta{Name: "obj4", Namespace: "acme-devel"}}
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "acme-devel")
		require.NoError(t, Create(ctx, s.client, "/sensu.io/generic/acme-devel/obj4", "acme-devel", obj4))

		// We should have 1 object when listing keys under the default namespace
		list := []*GenericObject{}
		pred := &store.SelectionPredicate{}
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "default")
		err = List(ctx, s.client, getGenericObjectsPath, &list, pred)
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Empty(t, pred.Continue)

		// Make sure the annonations & labels were initialized if nil
		assert.Equal(t, map[string]string{}, list[0].ObjectMeta.Annotations)
		assert.Equal(t, map[string]string{}, list[0].ObjectMeta.Labels)

		// We should have 2 objects when listing keys under the acme namespace
		list = []*GenericObject{}
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "acme")
		err = List(ctx, s.client, getGenericObjectsPath, &list, pred)
		require.NoError(t, err)
		assert.Len(t, list, 2)
		assert.Empty(t, pred.Continue)

		// We should have 1 object when listing keys under the acme-devel namespace
		list = []*GenericObject{}
		pred = &store.SelectionPredicate{}
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "acme-devel")
		err = List(ctx, s.client, getGenericObjectsPath, &list, pred)
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Empty(t, pred.Continue)

		// We should have 4 objects when listing through all namespaces
		list = []*GenericObject{}
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "")
		err = List(ctx, s.client, getGenericObjectsPath, &list, pred)
		require.NoError(t, err)
		assert.Len(t, list, 4)
		assert.Empty(t, pred.Continue)
	})
}

func TestListPagination(t *testing.T) {
	testWithEtcdStore(t, func(store *Store) {
		// Create a "testing" namespace in the store
		if err := store.CreateNamespace(context.Background(), types.FixtureNamespace("testing")); err != nil {
			t.Fatal(err)
		}

		// Add 42 objects in the store: 21 in the "default" namespace and 21 in
		// the "testing" namespace
		for i := 1; i <= 21; i++ {
			// We force the object name to be 2 digits "wide" in order to
			// have a "natural" lexicographic order: 01, 02, ... instead of 1,
			// 11, ...
			objectName := fmt.Sprintf("%.2d", i)
			object := &GenericObject{ObjectMeta: corev2.ObjectMeta{Name: objectName, Namespace: "default"}}

			ctx := context.WithValue(context.Background(), types.NamespaceKey, "default")
			if err := Create(ctx, store.client, getGenericObjectPath(object), "default", object); err != nil {
				t.Fatal(err)
			}

			object.Namespace = "testing"
			ctx = context.WithValue(context.Background(), types.NamespaceKey, "testing")
			if err := Create(ctx, store.client, getGenericObjectPath(object), "testing", object); err != nil {
				t.Fatal(err)
			}
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")
		t.Run("objects in default namespace", func(t *testing.T) {
			testListPagination(t, ctx, store, 10, 21)
		})

		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "testing")
		t.Run("objects in testing namespace", func(t *testing.T) {
			testListPagination(t, ctx, store, 10, 21)
		})

		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")
		t.Run("page size equals one", func(t *testing.T) {
			testListPagination(t, ctx, store, 1, 21)
		})

		ctx = context.Background()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")
		t.Run("page size bigger than set size", func(t *testing.T) {
			testListPagination(t, ctx, store, 1337, 21)
		})
	})
}

func testListPagination(t *testing.T, ctx context.Context, s *Store, limit, setSize int) {
	pred := &store.SelectionPredicate{Limit: int64(limit)}
	nFullPages := setSize / limit
	nLeftovers := setSize % limit

	for i := 0; i < nFullPages; i++ {
		objects := []*GenericObject{}
		err := List(ctx, s.client, getGenericObjectsPath, &objects, pred)
		if err != nil {
			t.Fatal(err)
		}

		if len(objects) != limit {
			t.Fatalf("Expected page %d to have %d objects but got %d", i, limit, len(objects))
		}

		offset := i * limit
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
		objects := []*GenericObject{}
		err := List(ctx, s.client, getGenericObjectsPath, &objects, pred)
		if err != nil {
			t.Fatal(err)
		}

		if len(objects) != nLeftovers {
			t.Fatalf("Expected last page with %d objects, got %d", nLeftovers, len(objects))
		}

		if pred.Continue != "" {
			t.Fatalf("Expected next continue token to be \"\", got %s", pred.Continue)
		}

		offset := limit * nFullPages
		for j, object := range objects {
			n := ((offset + j) % setSize) + 1
			expected := fmt.Sprintf("%.2d", n)

			if object.Name != expected {
				t.Fatalf("Expected %s, got %s", expected, object.Name)
			}
		}
	}
}

func TestUpdate(t *testing.T) {
	testWithEtcdStore(t, func(store *Store) {
		// Updating a non-existent object should fail
		obj := &GenericObject{Revision: 1}
		ctx := context.WithValue(context.Background(), types.NamespaceKey, "default")
		require.Error(t, Update(ctx, store.client, "/default/foo", "default", obj))

		// Create it first
		require.NoError(t, Create(ctx, store.client, "/default/foo", "default", obj))

		// Now make sure it gets properly updated
		obj.Revision = 2
		require.NoError(t, Update(ctx, store.client, "/default/foo", "default", obj))
		result := &GenericObject{}
		require.NoError(t, Get(ctx, store.client, "/default/foo", result))
		assert.Equal(t, uint32(2), obj.Revision)
	})
}

func TestUpdateWithValue(t *testing.T) {
	testWithEtcdStore(t, func(store *Store) {
		obj := &GenericObject{Revision: 1}
		b, err := proto.Marshal(obj)
		if err != nil {
			t.Fatalf("could not marshal the generic object: %s", err)
		}

		// Updating a non-existent object should fail
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, "default")
		valueComparison := kvc.KeyHasValue("/default/foo", b)
		require.Error(t, UpdateWithComparisons(ctx, store.client, "/default/foo", obj, valueComparison))

		// Create it first
		require.NoError(t, Create(ctx, store.client, "/default/foo", "default", obj))

		obj2 := &GenericObject{Revision: 2}
		b2, err := proto.Marshal(obj2)
		if err != nil {
			t.Fatalf("could not marshal the generic object: %s", err)
		}

		// Now try to update it while simulating an update in the mean time
		valueComparison = kvc.KeyHasValue("/default/foo", b2)
		require.Error(t, UpdateWithComparisons(ctx, store.client, "/default/foo", obj, valueComparison))

		// Updating with the right version should work
		valueComparison = kvc.KeyHasValue("/default/foo", b)
		require.NoError(t, UpdateWithComparisons(ctx, store.client, "/default/foo", obj2, valueComparison))
	})
}

func TestCount(t *testing.T) {
	testWithEtcdStore(t, func(s *Store) {
		// Create a second namespace
		require.NoError(t, s.CreateNamespace(context.Background(), types.FixtureNamespace("acme")))

		// Create a bunch of keys everywhere
		obj1 := &GenericObject{ObjectMeta: corev2.ObjectMeta{Name: "obj1", Namespace: "default"}}
		ctx := context.WithValue(context.Background(), types.NamespaceKey, "default")
		require.NoError(t, Create(ctx, s.client, "/sensu.io/generic/default/obj1", "default", obj1))

		obj2 := &GenericObject{ObjectMeta: corev2.ObjectMeta{Name: "obj2", Namespace: "acme"}}
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "acme")
		require.NoError(t, Create(ctx, s.client, "/sensu.io/generic/acme/obj2", "acme", obj2))

		obj3 := &GenericObject{ObjectMeta: corev2.ObjectMeta{Name: "obj3", Namespace: "acme"}}
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "acme")
		require.NoError(t, Create(ctx, s.client, "/sensu.io/generic/acme/obj3", "acme", obj3))

		// We should have 1 object when listing keys under the default namespace
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "default")
		count, err := Count(ctx, s.client, getGenericObjectsPath(ctx, ""))
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// We should have 2 objects when listing keys under the acme namespace
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "acme")
		count, err = Count(ctx, s.client, getGenericObjectsPath(ctx, ""))
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		// We should have 3 objects when listing through all namespaces
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "")
		count, err = Count(ctx, s.client, getGenericObjectsPath(ctx, ""))
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})
}

func TestUnmarshal(t *testing.T) {
	resource := &GenericObject{Revision: 1}
	jsonResource, _ := json.Marshal(resource)
	protoResource, _ := json.Marshal(resource)

	tests := []struct {
		name    string
		data    []byte
		v       interface{}
		wantErr bool
	}{
		{
			name: "JSON data",
			data: jsonResource,
			v:    resource,
		},
		{
			name: "Protobuf data",
			data: protoResource,
			v:    resource,
		},
		{
			name:    "Invalid serialized object",
			data:    protoResource,
			v:       string("foo"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := unmarshal(tt.data, tt.v); (err != nil) != tt.wantErr {
				t.Errorf("unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Equal(t, resource, tt.v)
			}
		})
	}
}
