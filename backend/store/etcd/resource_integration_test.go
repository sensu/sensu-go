//go:build integration && !race
// +build integration,!race

package etcd

import (
	"context"
	"reflect"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	"go.etcd.io/etcd/client/v3"
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
		patcher := &patch.Merge{MergePatch: []byte(`{"metadata":{"labels":{"42":"answer to life"}}}`)}
		if err := s.PatchResource(ctx, &patchedObj, "foo", patcher, nil); err != nil {
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

		// Determine the etag for the stored resource
		etag, err := store.ETag(storedObj)
		if err != nil {
			t.Fatalf("could not determine the etag: %s", err)
		}

		// An etag in a If-Match that does not match should return a precondition
		// error
		condition := &store.ETagCondition{
			IfMatch: `"12345"`,
		}
		err = s.PatchResource(ctx, &patchedObj, "foo", patcher, condition)
		if _, ok := err.(*store.ErrPreconditionFailed); !ok {
			t.Fatal("expected an error of type *store.ErrPreconditionFailed")
		}

		// A matching etag in a If-Match should proceed
		condition.IfMatch = etag
		if err = s.PatchResource(ctx, &patchedObj, "foo", patcher, condition); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		// A matching etag in a If-None-Match should return a precondition error
		condition.IfMatch = ""
		condition.IfNoneMatch = etag
		err = s.PatchResource(ctx, &patchedObj, "foo", patcher, condition)
		if _, ok := err.(*store.ErrPreconditionFailed); !ok {
			t.Fatal("expected an error of type *store.ErrPreconditionFailed")
		}

		// An etag in a If-None-Match that does not match should proceed
		condition.IfNoneMatch = `"12345"`
		if err = s.PatchResource(ctx, &patchedObj, "foo", patcher, condition); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})
}
