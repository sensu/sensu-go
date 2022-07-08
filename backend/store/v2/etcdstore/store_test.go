package etcdstore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
)

func fixtureTestResource(name string) *testResource {
	return &testResource{
		Metadata: &corev2.ObjectMeta{
			Namespace:   "default",
			Name:        name,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
	}
}

func TestCreateOrUpdate(t *testing.T) {
	testWithEtcdStore(t, func(s *etcdstore.Store) {
		// Create a namespace to work within
		ns := &corev2.Namespace{Name: "default"}
		ctx := context.Background()
		req := storev2.NewResourceRequestFromV2Resource(ns)
		wrapper, err := wrap.V2Resource(ns)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			t.Fatal(err)
		}
		// Create a resource under the default namespace
		fixture := fixtureTestResource("foo")
		req = storev2.NewResourceRequestFromResource(fixture)
		wrapper, err = wrap.Resource(fixture)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			t.Error(err)
		}
		// Repeating the call to the store should succeed
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			t.Error(err)
		}
		// A resource under an uncreated namespace should fail to create
		fixture.Metadata.Namespace = "notdefault"
		req = storev2.NewResourceRequestFromResource(fixture)
		wrapper, err = wrap.Resource(fixture)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err == nil {
			t.Error("expected non-nil error")
		} else if _, ok := err.(*store.ErrNamespaceMissing); !ok {
			t.Errorf("wrong error: %s", err)
		}
	})
}

func TestUpdateIfExists(t *testing.T) {
	testWithEtcdStore(t, func(s *etcdstore.Store) {
		// Create a namespace to work within
		ns := &corev2.Namespace{Name: "default"}
		ctx := context.Background()
		req := storev2.NewResourceRequestFromV2Resource(ns)
		wrapper, err := wrap.V2Resource(ns)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			t.Fatal(err)
		}
		// Create a resource under the default namespace
		fixture := fixtureTestResource("foo")
		req = storev2.NewResourceRequestFromResource(fixture)
		wrapper, err = wrap.Resource(fixture)
		if err != nil {
			t.Fatal(err)
		}
		// UpdateIfExists should fail
		if err := s.UpdateIfExists(ctx, req, wrapper); err == nil {
			t.Error("expected non-nil error")
		} else {
			if _, ok := err.(*store.ErrNotFound); !ok {
				t.Errorf("wrong error: %s", err)
			}
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			t.Fatal(err)
		}
		// UpdateIfExists should succeed
		if err := s.UpdateIfExists(ctx, req, wrapper); err != nil {
			t.Error(err)
		}
	})
}

func TestCreateIfNotExists(t *testing.T) {
	testWithEtcdStore(t, func(s *etcdstore.Store) {
		// Create a namespace to work within
		ns := &corev2.Namespace{Name: "default"}
		ctx := context.Background()
		req := storev2.NewResourceRequestFromV2Resource(ns)
		wrapper, err := wrap.V2Resource(ns)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			t.Fatal(err)
		}
		// Create a resource under the default namespace
		fixture := fixtureTestResource("foo")
		req = storev2.NewResourceRequestFromResource(fixture)
		wrapper, err = wrap.Resource(fixture)
		if err != nil {
			t.Fatal(err)
		}
		// CreateIfNotExists should succeed
		if err := s.CreateIfNotExists(ctx, req, wrapper); err != nil {
			t.Fatal(err)
		}
		// CreateIfNotExists should fail
		if err := s.CreateIfNotExists(ctx, req, wrapper); err == nil {
			t.Error("expected non-nil error")
		} else if _, ok := err.(*store.ErrAlreadyExists); !ok {
			t.Errorf("wrong error: %s", err)
		}
		// UpdateIfExists should succeed
		if err := s.UpdateIfExists(ctx, req, wrapper); err != nil {
			t.Error(err)
		}
		req.Namespace = "notexists"
		if err := s.CreateIfNotExists(ctx, req, wrapper); err == nil {
			t.Error("expected non-nil error")
		} else if _, ok := err.(*store.ErrNamespaceMissing); !ok {
			t.Errorf("expected ErrNamespaceMissing, got %T", err)
		}
	})
}

func TestGet(t *testing.T) {
	testWithEtcdStore(t, func(s *etcdstore.Store) {
		// Create a namespace to work within
		ns := &corev2.Namespace{Name: "default"}
		ctx := context.Background()
		req := storev2.NewResourceRequestFromV2Resource(ns)
		wrapper, err := wrap.V2Resource(ns)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			t.Fatal(err)
		}
		// Create a resource under the default namespace
		fixture := fixtureTestResource("foo")
		req = storev2.NewResourceRequestFromResource(fixture)
		wrapper, err = wrap.Resource(fixture)
		if err != nil {
			t.Fatal(err)
		}
		// CreateIfNotExists should succeed
		if err := s.CreateIfNotExists(ctx, req, wrapper); err != nil {
			t.Fatal(err)
		}
		got, err := s.Get(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if want := wrapper; !proto.Equal(got.(proto.Message), wrapper) {
			t.Errorf("bad resource; got %v, want %v", got, want)
		}
	})
}

func TestDelete(t *testing.T) {
	testWithEtcdStore(t, func(s *etcdstore.Store) {
		// Create a namespace to work within
		ns := &corev2.Namespace{Name: "default"}
		ctx := context.Background()
		req := storev2.NewResourceRequestFromV2Resource(ns)
		wrapper, err := wrap.V2Resource(ns)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			t.Fatal(err)
		}
		// Create a resource under the default namespace
		fixture := fixtureTestResource("foo")
		req = storev2.NewResourceRequestFromResource(fixture)
		wrapper, err = wrap.Resource(fixture)
		if err != nil {
			t.Fatal(err)
		}
		// CreateIfNotExists should succeed
		if err := s.CreateIfNotExists(ctx, req, wrapper); err != nil {
			t.Fatal(err)
		}
		if err := s.Delete(ctx, req); err != nil {
			t.Fatal(err)
		}
		if err := s.Delete(ctx, req); err == nil {
			t.Error("expected non-nil error")
		} else if _, ok := err.(*store.ErrNotFound); !ok {
			t.Errorf("expected ErrNotFound: got %s", err)
		}
		if _, err := s.Get(ctx, req); err == nil {
			t.Error("expected non-nil error")
		} else if _, ok := err.(*store.ErrNotFound); !ok {
			t.Errorf("expected ErrNotFound: got %s", err)
		}
	})
}

func TestList(t *testing.T) {
	testWithEtcdStore(t, func(s *etcdstore.Store) {
		// Create a namespace to work within
		ns := &corev2.Namespace{Name: "default"}
		ctx := context.Background()
		req := storev2.NewResourceRequestFromV2Resource(ns)
		wrapper, err := wrap.V2Resource(ns)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 10; i++ {
			// create 10 resources
			fixture := fixtureTestResource(fmt.Sprintf("foo-%d", i))
			req = storev2.NewResourceRequestFromResource(fixture)
			wrapper, err = wrap.Resource(fixture)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.CreateIfNotExists(ctx, req, wrapper); err != nil {
				t.Fatal(err)
			}
		}
		req = storev2.NewResourceRequest(new(testResource).GetTypeMeta(), "default", "anything", new(testResource).StoreName())
		pred := &store.SelectionPredicate{Limit: 5}
		// Test listing with limit of 5
		list, err := s.List(ctx, req, pred)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := list.Len(), 5; got != want {
			t.Errorf("wrong number of items: got %d, want %d", got, want)
		}
		if got, want := pred.Continue, "foo-4\x00"; got != want {
			t.Errorf("bad continue token: got %q, want %q", got, want)
		}
		// get the rest of the list
		list, err = s.List(ctx, req, pred)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := list.Len(), 5; got != want {
			t.Errorf("wrong number of items: got %d, want %d", got, want)
		}
		if pred.Continue != "" {
			t.Error("expected empty continue token")
		}
		// Test listing from all namespaces
		req.Namespace = ""
		pred = &store.SelectionPredicate{Limit: 5}
		list, err = s.List(ctx, req, pred)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := list.Len(), 5; got != want {
			t.Errorf("wrong number of items: got %d, want %d", got, want)
		}
		if got, want := pred.Continue, "default/foo-4\x00"; got != want {
			t.Errorf("bad continue token: got %q, want %q", got, want)
		}
		// get the rest of the list
		list, err = s.List(ctx, req, pred)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := list.Len(), 5; got != want {
			t.Errorf("wrong number of items: got %d, want %d", got, want)
		}
		if pred.Continue != "" {
			t.Error("expected empty continue token")
		}
		// Test listing in descending order
		pred.Continue = ""
		pred.Limit = 5
		req.SortOrder = storev2.SortDescend
		list, err = s.List(ctx, req, pred)
		if err != nil {
			t.Fatal(err)
		}
		if got := list.Len(); got == 0 {
			t.Fatalf("wrong number of items: got %d, want > %d", got, 0)
		}
		firstObj, err := list.(wrap.List)[0].Unwrap()
		if err != nil {
			t.Fatal(err)
		}
		if got, want := firstObj.GetMetadata().Name, "foo-9"; got != want {
			t.Errorf("unexpected first item in list: got %s, want %s", got, want)
		}
		list, err = s.List(ctx, req, pred) // get second chunk
		if err != nil {
			t.Fatal(err)
		}
		if got := list.Len(); got == 0 {
			t.Fatalf("wrong number of items: got %d, want > %d", got, 0)
		}
		firstObj, err = list.(wrap.List)[0].Unwrap()
		if err != nil {
			t.Fatal(err)
		}
		if got, want := firstObj.GetMetadata().Name, "foo-4"; got != want {
			t.Errorf("unexpected first item in list: got %s, want %s", got, want)
		}
		// Test listing in ascending order
		pred.Continue = ""
		req.SortOrder = storev2.SortAscend
		list, err = s.List(ctx, req, pred)
		if err != nil {
			t.Fatal(err)
		}
		if got := list.Len(); got == 0 {
			t.Fatalf("wrong number of items: got %d, want > %d", got, 0)
		}
		firstObj, err = list.(wrap.List)[0].Unwrap()
		if err != nil {
			t.Fatal(err)
		}
		if got, want := firstObj.GetMetadata().Name, "foo-0"; got != want {
			t.Errorf("unexpected first item in list: got %s, want %s", got, want)
		}
		list, err = s.List(ctx, req, pred) // get second chunk
		if err != nil {
			t.Fatal(err)
		}
		if got := list.Len(); got == 0 {
			t.Fatalf("wrong number of items: got %d, want > %d", got, 0)
		}
		firstObj, err = list.(wrap.List)[0].Unwrap()
		if err != nil {
			t.Fatal(err)
		}
		if got, want := firstObj.GetMetadata().Name, "foo-5"; got != want {
			t.Errorf("unexpected first item in list: got %s, want %s", got, want)
		}
	})
}

func TestExists(t *testing.T) {
	testWithEtcdStore(t, func(s *etcdstore.Store) {
		// Create a namespace to work within
		ns := &corev2.Namespace{Name: "default"}
		ctx := context.Background()
		req := storev2.NewResourceRequestFromV2Resource(ns)
		wrapper, err := wrap.V2Resource(ns)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
			t.Fatal(err)
		}
		fixture := fixtureTestResource("foo")
		req = storev2.NewResourceRequestFromResource(fixture)

		// Exists should return false
		got, err := s.Exists(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if want := false; got != want {
			t.Errorf("got true, want false")
		}

		// Create a resource under the default namespace
		wrapper, err = wrap.Resource(fixture)
		if err != nil {
			t.Fatal(err)
		}
		// CreateIfNotExists should succeed
		if err := s.CreateIfNotExists(ctx, req, wrapper); err != nil {
			t.Fatal(err)
		}
		got, err = s.Exists(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if want := true; got != want {
			t.Errorf("got false, want true")
		}
	})

}
