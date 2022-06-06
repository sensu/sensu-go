package postgres

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

func testWithPostgresStoreV2(t *testing.T, fn func(storev2.Interface)) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping postgres test")
		return
	}
	pgURL := os.Getenv("PG_URL")
	if pgURL == "" {
		t.Skip("skipping postgres test")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, err := pgxpool.Connect(ctx, pgURL)
	if err != nil {
		t.Fatal(err)
	}
	dbName := "sensu" + strings.ReplaceAll(uuid.New().String(), "-", "")
	if _, err := db.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s;", dbName)); err != nil {
		t.Fatal(err)
	}
	defer dropAll(context.Background(), dbName, pgURL)
	db.Close()
	db, err = pgxpool.Connect(ctx, fmt.Sprintf("dbname=%s ", dbName)+pgURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := upgrade(ctx, db); err != nil {
		t.Fatal(err)
	}
	fn(NewStoreV2(db, nil))
}

func TestStoreCreateOrUpdate(t *testing.T) {
	testWithPostgresStoreV2(t, func(s storev2.Interface) {
		cfg := corev3.FixtureEntityConfig("foo")
		ctx := context.Background()
		req := storev2.NewResourceRequestFromResource(ctx, cfg)
		req.UsePostgres = true
		wrapper := WrapEntityConfig(cfg)
		if err := s.CreateOrUpdate(req, wrapper); err != nil {
			t.Error(err)
		}
		// Repeating the call to the store should succeed
		if err := s.CreateOrUpdate(req, wrapper); err != nil {
			t.Error(err)
		}
		rows, err := s.(*StoreV2).db.Query(context.Background(), "SELECT * FROM entity_configs")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		rowCount := 0
		for rows.Next() {
			rowCount++
		}
		if got, want := rowCount, 1; got != want {
			t.Errorf("bad row count: got %d, want %d", got, want)
		}
	})
}

func TestStoreUpdateIfExists(t *testing.T) {
	testWithPostgresStoreV2(t, func(s storev2.Interface) {
		cfg := corev3.FixtureEntityConfig("foo")
		ctx := context.Background()
		req := storev2.NewResourceRequestFromResource(ctx, cfg)
		req.UsePostgres = true
		wrapper := WrapEntityConfig(cfg)
		// UpdateIfExists should fail
		if err := s.UpdateIfExists(req, wrapper); err == nil {
			t.Error("expected non-nil error")
		} else {
			if _, ok := err.(*store.ErrNotFound); !ok {
				t.Errorf("wrong error: %s", err)
			}
		}
		if err := s.CreateOrUpdate(req, wrapper); err != nil {
			t.Fatal(err)
		}
		// UpdateIfExists should succeed
		if err := s.UpdateIfExists(req, wrapper); err != nil {
			t.Error(err)
		}
	})
}

func TestStoreCreateIfNotExists(t *testing.T) {
	testWithPostgresStoreV2(t, func(s storev2.Interface) {
		cfg := corev3.FixtureEntityConfig("foo")
		ctx := context.Background()
		req := storev2.NewResourceRequestFromResource(ctx, cfg)
		req.UsePostgres = true
		wrapper := WrapEntityConfig(cfg)
		// CreateIfNotExists should succeed
		if err := s.CreateIfNotExists(req, wrapper); err != nil {
			t.Fatal(err)
		}
		// CreateIfNotExists should fail
		if err := s.CreateIfNotExists(req, wrapper); err == nil {
			t.Error("expected non-nil error")
		} else if _, ok := err.(*store.ErrAlreadyExists); !ok {
			t.Errorf("wrong error: %s", err)
		}
		// UpdateIfExists should succeed
		if err := s.UpdateIfExists(req, wrapper); err != nil {
			t.Error(err)
		}
	})
}

func TestStoreGet(t *testing.T) {
	testWithPostgresStoreV2(t, func(s storev2.Interface) {
		cfg := corev3.FixtureEntityConfig("foo")
		ctx := context.Background()
		req := storev2.NewResourceRequestFromResource(ctx, cfg)
		req.UsePostgres = true
		wrapper := WrapEntityConfig(cfg)
		// CreateIfNotExists should succeed
		if err := s.CreateOrUpdate(req, wrapper); err != nil {
			t.Fatal(err)
		}
		got, err := s.Get(req)
		if err != nil {
			t.Fatal(err)
		}
		if want := wrapper; !reflect.DeepEqual(got, wrapper) {
			t.Errorf("bad resource; got %#v, want %#v", got, want)
		}
	})
}

func TestStoreDelete(t *testing.T) {
	testWithPostgresStoreV2(t, func(s storev2.Interface) {
		cfg := corev3.FixtureEntityConfig("foo")
		ctx := context.Background()
		req := storev2.NewResourceRequestFromResource(ctx, cfg)
		req.UsePostgres = true
		wrapper := WrapEntityConfig(cfg)
		// CreateIfNotExists should succeed
		if err := s.CreateIfNotExists(req, wrapper); err != nil {
			t.Fatal(err)
		}
		if err := s.Delete(req); err != nil {
			t.Fatal(err)
		}
		if err := s.Delete(req); err == nil {
			t.Error("expected non-nil error")
		} else if _, ok := err.(*store.ErrNotFound); !ok {
			t.Errorf("expected ErrNotFound: got %s", err)
		}
		if _, err := s.Get(req); err == nil {
			t.Error("expected non-nil error")
		} else if _, ok := err.(*store.ErrNotFound); !ok {
			t.Errorf("expected ErrNotFound: got %s", err)
		}
	})
}

func TestStoreList(t *testing.T) {
	testWithPostgresStoreV2(t, func(s storev2.Interface) {
		for i := 0; i < 10; i++ {
			// create 10 resources
			cfg := corev3.FixtureEntityConfig(fmt.Sprintf("foo-%d", i))
			ctx := context.Background()
			req := storev2.NewResourceRequestFromResource(ctx, cfg)
			req.UsePostgres = true
			wrapper := WrapEntityConfig(cfg)
			if err := s.CreateIfNotExists(req, wrapper); err != nil {
				t.Fatal(err)
			}
		}
		ctx := context.Background()
		req := storev2.NewResourceRequest(ctx, "default", "anything", new(corev3.EntityConfig).StoreName())
		req.UsePostgres = true
		pred := &store.SelectionPredicate{Limit: 5}
		// Test listing with limit of 5
		list, err := s.List(req, pred)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := list.Len(), 5; got != want {
			t.Errorf("wrong number of items: got %d, want %d", got, want)
		}
		if got, want := pred.Continue, `{"offset":5}`; got != want {
			t.Errorf("bad continue token: got %q, want %q", got, want)
		}
		// get the rest of the list
		pred.Limit = 6
		list, err = s.List(req, pred)
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
		list, err = s.List(req, pred)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := list.Len(), 5; got != want {
			t.Errorf("wrong number of items: got %d, want %d", got, want)
		}
		if got, want := pred.Continue, `{"offset":5}`; got != want {
			t.Errorf("bad continue token: got %q, want %q", got, want)
		}
		pred.Limit = 6
		// get the rest of the list
		list, err = s.List(req, pred)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := list.Len(), 5; got != want {
			t.Errorf("wrong number of items: got %d, want %d", got, want)
		}
		if pred.Continue != "" {
			t.Error("expected empty continue token")
		}
		pred.Limit = 5
		// Test listing in descending order
		pred.Continue = ""
		req.SortOrder = storev2.SortDescend
		list, err = s.List(req, pred)
		if err != nil {
			t.Fatal(err)
		}
		if got := list.Len(); got == 0 {
			t.Fatalf("wrong number of items: got %d, want > %d", got, 0)
		}
		firstObj, err := list.(WrapList)[0].Unwrap()
		if err != nil {
			t.Fatal(err)
		}
		if got, want := firstObj.GetMetadata().Name, "foo-9"; got != want {
			t.Errorf("unexpected first item in list: got %s, want %s", got, want)
		}
		// Test listing in ascending order
		pred.Continue = ""
		req.SortOrder = storev2.SortAscend
		list, err = s.List(req, pred)
		if err != nil {
			t.Fatal(err)
		}
		if got := list.Len(); got == 0 {
			t.Fatalf("wrong number of items: got %d, want > %d", got, 0)
		}
		firstObj, err = list.(WrapList)[0].Unwrap()
		if err != nil {
			t.Fatal(err)
		}
		if got, want := firstObj.GetMetadata().Name, "foo-0"; got != want {
			t.Errorf("unexpected first item in list: got %s, want %s", got, want)
		}
	})
}

func TestStoreExists(t *testing.T) {
	testWithPostgresStoreV2(t, func(s storev2.Interface) {
		cfg := corev3.FixtureEntityConfig("foo")
		ctx := context.Background()
		req := storev2.NewResourceRequestFromResource(ctx, cfg)
		req.UsePostgres = true
		// Exists should return false
		got, err := s.Exists(req)
		if err != nil {
			t.Fatal(err)
		}
		if want := false; got != want {
			t.Errorf("got true, want false")
		}

		// Create a resource under the default namespace
		wrapper := WrapEntityConfig(cfg)
		// CreateIfNotExists should succeed
		if err := s.CreateIfNotExists(req, wrapper); err != nil {
			t.Fatal(err)
		}
		got, err = s.Exists(req)
		if err != nil {
			t.Fatal(err)
		}
		if want := true; got != want {
			t.Errorf("got false, want true")
		}
	})

}

func TestStorePatch(t *testing.T) {
	testWithPostgresStoreV2(t, func(s storev2.Interface) {
		cfg := corev3.FixtureEntityConfig("foo")
		ctx := context.Background()
		req := storev2.NewResourceRequestFromResource(ctx, cfg)
		req.UsePostgres = true
		wrapper := WrapEntityConfig(cfg)
		if err := s.CreateOrUpdate(req, wrapper); err != nil {
			t.Error(err)
		}
		patcher := &patch.Merge{
			MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
		}

		if err := s.Patch(req, wrapper, patcher, nil); err != nil {
			t.Fatal(err)
		}

		updatedWrapper, err := s.Get(req)
		if err != nil {
			t.Fatal(err)
		}

		updated, err := updatedWrapper.Unwrap()
		if err != nil {
			t.Fatal(err)
		}

		if got, want := updated.GetMetadata().Labels["food"], "hummus"; got != want {
			t.Errorf("bad patched labels: got %q, want %q", got, want)
		}
	})
}

func TestStoreGetMultiple(t *testing.T) {
	testWithPostgresStoreV2(t, func(s storev2.Interface) {
		reqs := make([]storev2.ResourceRequest, 0)
		for i := 0; i < 10; i++ {
			// create 10 resources
			cfg := corev3.FixtureEntityConfig(fmt.Sprintf("foo-%d", i))
			ctx := context.Background()
			req := storev2.NewResourceRequestFromResource(ctx, cfg)
			req.UsePostgres = true
			reqs = append(reqs, req)
			wrapper := WrapEntityConfig(cfg)
			if err := s.CreateIfNotExists(req, wrapper); err != nil {
				t.Fatal(err)
			}
		}
		result, err := s.(*StoreV2).GetMultiple(context.Background(), reqs[:5])
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(result), 5; got != want {
			t.Fatalf("bad number of results: got %d, want %d", got, want)
		}
		for i := 0; i < 5; i++ {
			wrapper, ok := result[reqs[i]]
			if !ok {
				t.Errorf("missing result %d", i)
				continue
			}
			var cfg corev3.EntityConfig
			if err := wrapper.UnwrapInto(&cfg); err != nil {
				t.Error(err)
				continue
			}
			if got, want := len(cfg.Subscriptions), 2; got != want {
				t.Errorf("wrong number of subscriptions, got = %v, want %v", got, want)
			}
			if got, want := len(cfg.KeepaliveHandlers), 1; got != want {
				t.Errorf("wrong number of keepalive handlers, got = %v, want %v", got, want)
			}
		}
		req := reqs[0]
		req.Namespace = "notexists"
		result, err = s.(*StoreV2).GetMultiple(context.Background(), []storev2.ResourceRequest{req})
		if err != nil {
			t.Fatal(err)
		}
		if len(result) != 0 {
			t.Fatal("wrong result length")
		}
	})
}
