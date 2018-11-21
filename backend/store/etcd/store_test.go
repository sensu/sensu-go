// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/stretchr/testify/assert"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/require"
)

func testWithEtcd(t *testing.T, f func(store.Store)) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)

	s := NewStore(client, e.Name())

	// Mock a default namespace
	require.NoError(t, s.CreateNamespace(context.Background(), types.FixtureNamespace("default")))

	f(s)
}

func testWithEtcdClient(t *testing.T, f func(store.Store, *clientv3.Client)) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)

	s := NewStore(client, e.Name())

	// Mock a default namespace
	require.NoError(t, s.CreateNamespace(context.Background(), types.FixtureNamespace("default")))

	f(s)
}

type genericObject struct {
	Namespace string
	Name      string
	Revision  int
}

func (g *genericObject) GetNamespace() string {
	return g.Namespace
}

func TestCreate(t *testing.T) {
	testWithEtcdStore(t, func(store *Store) {
		// Creating a namespaced key that does not exist should work
		obj := &genericObject{}
		ctx := context.WithValue(context.Background(), types.NamespaceKey, "default")
		err := store.create(ctx, "/default/foo", "default", obj)
		assert.NoError(t, err)

		// Creating this same key should return an error
		err = store.create(ctx, "/default/foo", "default", obj)
		assert.Error(t, err)

		// We should also be able to create a global object
		err = store.create(ctx, "/foo", "", obj)
		assert.NoError(t, err)
	})
}

func TestCreateOrUpdate(t *testing.T) {
	testWithEtcdStore(t, func(store *Store) {
		// Creating a namespaced key that does not exist should work
		obj := &genericObject{Revision: 1}
		ctx := context.WithValue(context.Background(), types.NamespaceKey, "default")
		err := store.createOrUpdate(ctx, "/default/foo", "default", obj)
		assert.NoError(t, err)

		// Creating this same key should also work, but the revision should be
		// different
		obj2 := &genericObject{Revision: 2}
		err = store.createOrUpdate(ctx, "/default/foo", "default", obj)
		assert.NoError(t, err)
		result := &genericObject{}
		err = store.get(ctx, "/default/foo", obj2)
		assert.NoError(t, err)
		assert.NotEqual(t, obj.Revision, result.Revision)

		// We should also be able to create a global object
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "")
		err = store.createOrUpdate(ctx, "/foo", "", obj)
		assert.NoError(t, err)
	})
}

func TestDelete(t *testing.T) {
	testWithEtcdStore(t, func(store *Store) {
		// Deleting a non-existant key should fail
		ctx := context.WithValue(context.Background(), types.NamespaceKey, "default")
		require.Error(t, store.delete(ctx, "/default/foo"))

		// Create it first
		obj := &genericObject{}
		require.NoError(t, store.create(ctx, "/default/foo", "default", obj))

		// Now make sure it gets properly deleted
		require.NoError(t, store.delete(ctx, "/default/foo"))
		result := &genericObject{}
		require.Error(t, store.get(ctx, "/default/foo", result))
	})
}
func TestGet(t *testing.T) {
	testWithEtcdStore(t, func(store *Store) {
		// Create a namespaced key
		obj := &genericObject{Revision: 1}
		ctx := context.WithValue(context.Background(), types.NamespaceKey, "default")
		err := store.create(ctx, "/default/foo", "default", obj)
		assert.NoError(t, err)

		// Retrieve the namespaced key and make sure it's the expected object
		result := &genericObject{}
		err = store.get(ctx, "/default/foo", result)
		assert.NoError(t, err)
		assert.Equal(t, obj, result)

		// Create a global key
		obj2 := &genericObject{Revision: 2}
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "")
		err = store.create(ctx, "/foo", "", obj2)
		assert.NoError(t, err)

		// Retrieve the global key and make sure it's the expected object
		result2 := &genericObject{}
		err = store.get(ctx, "/foo", result2)
		assert.NoError(t, err)
		assert.Equal(t, obj2, result2)
	})
}

var genericKeyBuilder = store.NewKeyBuilder("generic")

func getGenericObjectPath(obj *genericObject) string {
	return genericKeyBuilder.WithResource(obj).Build(obj.Name)
}

func getGenericObjectsPath(ctx context.Context, name string) string {
	return genericKeyBuilder.WithContext(ctx).Build(name)
}
func TestList(t *testing.T) {
	testWithEtcdStore(t, func(store *Store) {
		// Create a second namespace
		require.NoError(t, store.CreateNamespace(context.Background(), types.FixtureNamespace("acme")))

		// Create a bunch of keys everywhere
		obj1 := &genericObject{Name: "obj1", Namespace: "default"}
		ctx := context.WithValue(context.Background(), types.NamespaceKey, "default")
		require.NoError(t, store.create(ctx, "/sensu.io/generic/default/obj1", "default", obj1))

		obj2 := &genericObject{Name: "obj2", Namespace: "acme"}
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "acme")
		require.NoError(t, store.create(ctx, "/sensu.io/generic/acme/obj2", "acme", obj2))

		obj3 := &genericObject{Name: "obj3", Namespace: "acme"}
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "acme")
		require.NoError(t, store.create(ctx, "/sensu.io/generic/acme/obj3", "acme", obj3))

		// We should have 1 object when listing keys under the default namespace
		list := []*genericObject{}
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "default")
		require.NoError(t, store.list(ctx, getGenericObjectsPath, &list))
		assert.Len(t, list, 1)

		// We should have 2 objects when listing keys under the acme namespace
		list = []*genericObject{}
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "acme")
		require.NoError(t, store.list(ctx, getGenericObjectsPath, &list))
		assert.Len(t, list, 2)

		// We should have 3 objects when listing through all namespaces
		list = []*genericObject{}
		ctx = context.WithValue(context.Background(), types.NamespaceKey, "*")
		require.NoError(t, store.list(ctx, getGenericObjectsPath, &list))
		assert.Len(t, list, 3)
	})
}

func TestUpdate(t *testing.T) {
	testWithEtcdStore(t, func(store *Store) {
		// Updating a non-existent object should fail
		obj := &genericObject{Revision: 1}
		ctx := context.WithValue(context.Background(), types.NamespaceKey, "default")
		require.Error(t, store.update(ctx, "/default/foo", "default", obj))

		// Create it first
		require.NoError(t, store.create(ctx, "/default/foo", "default", obj))

		// Now make sure it gets properly updated
		obj.Revision = 2
		require.NoError(t, store.update(ctx, "/default/foo", "default", obj))
		result := &genericObject{}
		require.NoError(t, store.get(ctx, "/default/foo", result))
		assert.Equal(t, 2, obj.Revision)
	})
}
