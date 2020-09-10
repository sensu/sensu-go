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
		obj := &GenericObject{Revision: 42, ObjectMeta: corev2.ObjectMeta{Name: "foo", Namespace: "default"}}
		ctx := context.WithValue(context.Background(), corev2.NamespaceKey, "default")
		if err := s.CreateOrUpdateResource(ctx, obj); err != nil {
			t.Fatalf("could not create a resource: %s", err)
		}

		// Patch the resource
		patchedObj := GenericObject{}
		patcher := &patch.Merge{JSONPatch: []byte(`{"metadata":{"labels":{"42":"answer to life"}}}`)}
		_, err := s.PatchResource(ctx, &patchedObj, "foo", patcher, []byte{})
		if err != nil {
			t.Fatalf("could not apply the patch: %s", err)
		}

		// Make sure the stored and patched resources are the same
		storedObj := GenericObject{}
		if err := s.GetResource(ctx, "foo", &storedObj); err != nil {
			t.Fatalf("could not retrieve the stored resource: %s", err)
		}
		if !reflect.DeepEqual(patchedObj, storedObj) {
			t.Errorf("Store.PatchResource() = %#v, want %#v", patchedObj, storedObj)
		}
	})
}
