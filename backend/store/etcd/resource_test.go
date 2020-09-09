// +build integration,!race

package etcd

import (
	"context"
	"reflect"
	"testing"

	"github.com/coreos/etcd/clientv3"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
)

func TestStore_PatchResource(t *testing.T) {
	testWithEtcdClient(t, func(s store.Store, client *clientv3.Client) {
		// Create a resource
		obj := &GenericObject{Revision: 42}
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, "default")
		if err := Create(ctx, client, "/default/foo", "default", obj); err != nil {
			t.Fatalf("could not create a resource: %s", err)
		}

		// Patch the resource
		patchedObj := GenericObject{}
		patcher := &patch.Merge{JSONPatch: []byte(`{"metadata":{"name":"answer to life"}}`)}
		_, err := s.PatchResource(ctx, &patchedObj, "/default/foo", patcher, []byte{})
		if err != nil {
			t.Fatalf("could not apply the patch: %s", err)
		}

		// Make sure the stored and patched resources are the same
		storedObj := GenericObject{}
		if err := Get(ctx, client, "/default/foo", &storedObj); err != nil {
			t.Fatalf("could not retrieve the stored resource: %s", err)
		}
		if !reflect.DeepEqual(patchedObj, storedObj) {
			t.Errorf("Store.PatchResource() = %#v, want %#v", patchedObj, storedObj)
		}
	})
}
