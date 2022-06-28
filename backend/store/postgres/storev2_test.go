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
	corev2 "github.com/sensu/sensu-go/api/core/v2"
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

func testCreateNamespace(t *testing.T, s storev2.Interface, name string) {
	t.Helper()
	namespace := corev3.FixtureNamespace(name)
	req := storev2.NewResourceRequestFromResource(namespace)
	req.UsePostgres = true
	wrapper := WrapNamespace(namespace)
	if err := s.CreateOrUpdate(context.Background(), req, wrapper); err != nil {
		t.Error(err)
	}
}

func testCreateEntityConfig(t *testing.T, s storev2.Interface, name string) {
	t.Helper()
	ctx := context.Background()
	cfg := corev3.FixtureEntityConfig(name)
	req := storev2.NewResourceRequestFromResource(cfg)
	req.UsePostgres = true
	wrapper := WrapEntityConfig(cfg)
	if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
		t.Error(err)
	}
}

func testCreateEntityState(t *testing.T, s storev2.Interface, name string) {
	t.Helper()
	ctx := context.Background()
	state := corev3.FixtureEntityState(name)
	req := storev2.NewResourceRequestFromResource(state)
	req.UsePostgres = true
	wrapper := WrapEntityState(state)
	if err := s.CreateOrUpdate(ctx, req, wrapper); err != nil {
		t.Error(err)
	}
}

func TestStoreCreateOrUpdate(t *testing.T) {
	type args struct {
		req     storev2.ResourceRequest
		wrapper storev2.Wrapper
	}
	tests := []struct {
		name        string
		args        args
		verifyQuery string
		beforeHook  func(*testing.T, storev2.Interface)
		want        int
	}{
		{
			name: "entity configs can be created and updated",
			args: func() args {
				cfg := corev3.FixtureEntityConfig("foo")
				req := storev2.NewResourceRequestFromResource(cfg)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapEntityConfig(cfg),
				}
			}(),
			verifyQuery: fmt.Sprintf("SELECT * FROM %s", entityConfigStoreName),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
			},
			want: 1,
		},
		{
			name: "entity states can be created and updated",
			args: func() args {
				state := corev3.FixtureEntityState("foo")
				req := storev2.NewResourceRequestFromResource(state)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapEntityState(state),
				}
			}(),
			verifyQuery: fmt.Sprintf("SELECT * FROM %s", entityStateStoreName),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
				testCreateEntityConfig(t, s, "foo")
			},
			want: 1,
		},
		{
			name: "namespaces can be created and updated",
			args: func() args {
				ns := corev3.FixtureNamespace("bar")
				req := storev2.NewResourceRequestFromResource(ns)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapNamespace(ns),
				}
			}(),
			verifyQuery: fmt.Sprintf("SELECT * FROM %s", namespaceStoreName),
			want:        1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWithPostgresStoreV2(t, func(s storev2.Interface) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, s)
				}
				ctx := context.Background()
				if err := s.CreateOrUpdate(ctx, tt.args.req, tt.args.wrapper); err != nil {
					t.Error(err)
				}

				// Repeating the call to the store should succeed
				if err := s.CreateOrUpdate(ctx, tt.args.req, tt.args.wrapper); err != nil {
					t.Error(err)
				}
				rows, err := s.(*StoreV2).db.Query(context.Background(), tt.verifyQuery)
				if err != nil {
					t.Fatal(err)
				}
				defer rows.Close()
				got := 0
				for rows.Next() {
					got++
				}
				if got != tt.want {
					t.Errorf("bad row count: got %d, want %d", got, tt.want)
				}
			})
		})
	}
}

func TestStoreUpdateIfExists(t *testing.T) {
	type args struct {
		req     storev2.ResourceRequest
		wrapper storev2.Wrapper
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.Interface)
	}{
		{
			name: "entity configs can be updated if one exists",
			args: func() args {
				cfg := corev3.FixtureEntityConfig("foo")
				req := storev2.NewResourceRequestFromResource(cfg)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapEntityConfig(cfg),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
			},
		},
		{
			name: "entity states can be updated if one exists",
			args: func() args {
				state := corev3.FixtureEntityState("foo")
				req := storev2.NewResourceRequestFromResource(state)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapEntityState(state),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
				testCreateEntityConfig(t, s, "foo")
			},
		},
		{
			name: "namespaces can be updated if one exists",
			args: func() args {
				ns := corev3.FixtureNamespace("bar")
				req := storev2.NewResourceRequestFromResource(ns)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapNamespace(ns),
				}
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWithPostgresStoreV2(t, func(s storev2.Interface) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, s)
				}
				ctx := context.Background()

				// UpdateIfExists should fail
				if err := s.UpdateIfExists(ctx, tt.args.req, tt.args.wrapper); err == nil {
					t.Error("expected non-nil error")
				} else {
					if _, ok := err.(*store.ErrNotFound); !ok {
						t.Errorf("wrong error: %s", err)
					}
				}
				if err := s.CreateOrUpdate(ctx, tt.args.req, tt.args.wrapper); err != nil {
					t.Fatal(err)
				}

				// UpdateIfExists should succeed
				if err := s.UpdateIfExists(ctx, tt.args.req, tt.args.wrapper); err != nil {
					t.Error(err)
				}
			})
		})
	}
}

func TestStoreCreateIfNotExists(t *testing.T) {
	type args struct {
		req     storev2.ResourceRequest
		wrapper storev2.Wrapper
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.Interface)
	}{
		{
			name: "entity configs can be created if one does not exist",
			args: func() args {
				cfg := corev3.FixtureEntityConfig("foo")
				req := storev2.NewResourceRequestFromResource(cfg)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapEntityConfig(cfg),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
			},
		},
		{
			name: "entity states can be created if one does not exist",
			args: func() args {
				state := corev3.FixtureEntityState("foo")
				req := storev2.NewResourceRequestFromResource(state)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapEntityState(state),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
				testCreateEntityConfig(t, s, "foo")
			},
		},
		{
			name: "namespaces can be created if one does not exist",
			args: func() args {
				ns := corev3.FixtureNamespace("bar")
				req := storev2.NewResourceRequestFromResource(ns)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapNamespace(ns),
				}
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWithPostgresStoreV2(t, func(s storev2.Interface) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, s)
				}

				ctx := context.Background()

				// CreateIfNotExists should succeed
				if err := s.CreateIfNotExists(ctx, tt.args.req, tt.args.wrapper); err != nil {
					t.Fatal(err)
				}

				// CreateIfNotExists should fail
				if err := s.CreateIfNotExists(ctx, tt.args.req, tt.args.wrapper); err == nil {
					t.Error("expected non-nil error")
				} else if _, ok := err.(*store.ErrAlreadyExists); !ok {
					t.Errorf("wrong error: %s", err)
				}

				// UpdateIfExists should succeed
				if err := s.UpdateIfExists(ctx, tt.args.req, tt.args.wrapper); err != nil {
					t.Error(err)
				}
			})
		})
	}
}

func TestStoreGet(t *testing.T) {
	type args struct {
		req storev2.ResourceRequest
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.Interface)
		want       storev2.Wrapper
	}{
		{
			name: "an entity config can be retrieved",
			args: func() args {
				cfg := corev3.FixtureEntityConfig("foo")
				req := storev2.NewResourceRequestFromResource(cfg)
				req.UsePostgres = true
				return args{
					req: req,
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
				testCreateEntityConfig(t, s, "foo")
			},
			want: func() storev2.Wrapper {
				cfg := corev3.FixtureEntityConfig("foo")
				return WrapEntityConfig(cfg)
			}(),
		},
		{
			name: "an entity state can be retrieved",
			args: func() args {
				state := corev3.FixtureEntityState("foo")
				req := storev2.NewResourceRequestFromResource(state)
				req.UsePostgres = true
				return args{
					req: req,
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
				testCreateEntityConfig(t, s, "foo")
				testCreateEntityState(t, s, "foo")
			},
			want: func() storev2.Wrapper {
				state := corev3.FixtureEntityState("foo")
				return WrapEntityState(state)
			}(),
		},
		{
			name: "a namespace can be retrieved",
			args: func() args {
				ns := corev3.FixtureNamespace("foo")
				req := storev2.NewResourceRequestFromResource(ns)
				req.UsePostgres = true
				return args{
					req: req,
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "foo")
			},
			want: func() storev2.Wrapper {
				ns := corev3.FixtureNamespace("foo")
				return WrapNamespace(ns)
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWithPostgresStoreV2(t, func(s storev2.Interface) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, s)
				}

				got, err := s.Get(context.Background(), tt.args.req)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("bad resource; got %#v, want %#v", got, tt.want)
				}
			})
		})
	}
}

func TestStoreDelete(t *testing.T) {
	type args struct {
		req     storev2.ResourceRequest
		wrapper storev2.Wrapper
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.Interface)
	}{
		{
			name: "an entity config can be deleted",
			args: func() args {
				cfg := corev3.FixtureEntityConfig("foo")
				req := storev2.NewResourceRequestFromResource(cfg)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapEntityConfig(cfg),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
			},
		},
		{
			name: "an entity state can be deleted",
			args: func() args {
				state := corev3.FixtureEntityState("foo")
				req := storev2.NewResourceRequestFromResource(state)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapEntityState(state),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
				testCreateEntityConfig(t, s, "foo")
			},
		},
		{
			name: "a namespace can be deleted",
			args: func() args {
				ns := corev3.FixtureNamespace("bar")
				req := storev2.NewResourceRequestFromResource(ns)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapNamespace(ns),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWithPostgresStoreV2(t, func(s storev2.Interface) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, s)
				}
				ctx := context.Background()
				// CreateIfNotExists should succeed
				if err := s.CreateIfNotExists(ctx, tt.args.req, tt.args.wrapper); err != nil {
					t.Fatal(err)
				}
				if err := s.Delete(ctx, tt.args.req); err != nil {
					t.Fatal(err)
				}
				if err := s.Delete(ctx, tt.args.req); err == nil {
					t.Error("expected non-nil error")
				} else if _, ok := err.(*store.ErrNotFound); !ok {
					t.Errorf("expected ErrNotFound: got %s", err)
				}
				if _, err := s.Get(ctx, tt.args.req); err == nil {
					t.Error("expected non-nil error")
				} else if _, ok := err.(*store.ErrNotFound); !ok {
					t.Errorf("expected ErrNotFound: got %s", err)
				}
			})
		})
	}
}

func TestStoreList(t *testing.T) {
	type args struct {
		req  storev2.ResourceRequest
		pred *store.SelectionPredicate
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.Interface)
	}{
		{
			name: "entity configs can be listed",
			args: func() args {
				req := storev2.NewResourceRequest(corev2.TypeMeta{Type: "EntityConfig", APIVersion: "core/v3"}, "default", "anything", entityConfigStoreName)
				req.UsePostgres = true
				pred := &store.SelectionPredicate{Limit: 5}
				return args{
					req:  req,
					pred: pred,
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					testCreateEntityConfig(t, s, entityName)
				}
			},
		},
		{
			name: "entity states can be listed",
			args: func() args {
				req := storev2.NewResourceRequest(corev2.TypeMeta{Type: "EntityState", APIVersion: "core/v3"}, "default", "anything", entityStateStoreName)
				req.UsePostgres = true
				pred := &store.SelectionPredicate{Limit: 5}
				return args{
					req:  req,
					pred: pred,
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					testCreateEntityConfig(t, s, entityName)
					testCreateEntityState(t, s, entityName)
				}
			},
		},
		{
			name: "namespaces can be listed",
			args: func() args {
				req := storev2.NewResourceRequest(corev2.TypeMeta{Type: "Namespace", APIVersion: "core/v3"}, "", "anything", namespaceStoreName)
				req.UsePostgres = true
				pred := &store.SelectionPredicate{Limit: 5}
				return args{
					req:  req,
					pred: pred,
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				for i := 0; i < 10; i++ {
					namespaceName := fmt.Sprintf("foo-%d", i)
					testCreateNamespace(t, s, namespaceName)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWithPostgresStoreV2(t, func(s storev2.Interface) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, s)
				}

				ctx := context.Background()

				// Test listing with limit of 5
				list, err := s.List(ctx, tt.args.req, tt.args.pred)
				if err != nil {
					t.Fatal(err)
				}
				if got, want := list.Len(), 5; got != want {
					t.Errorf("wrong number of items: got %d, want %d", got, want)
				}
				if got, want := tt.args.pred.Continue, `{"offset":5}`; got != want {
					t.Errorf("bad continue token: got %q, want %q", got, want)
				}

				// get the rest of the list
				tt.args.pred.Limit = 6
				list, err = s.List(ctx, tt.args.req, tt.args.pred)
				if err != nil {
					t.Fatal(err)
				}
				if got, want := list.Len(), 5; got != want {
					t.Errorf("wrong number of items: got %d, want %d", got, want)
				}
				if tt.args.pred.Continue != "" {
					t.Error("expected empty continue token")
				}

				// Test listing from all namespaces
				tt.args.req.Namespace = ""
				tt.args.pred = &store.SelectionPredicate{Limit: 5}
				list, err = s.List(ctx, tt.args.req, tt.args.pred)
				if err != nil {
					t.Fatal(err)
				}
				if got, want := list.Len(), 5; got != want {
					t.Errorf("wrong number of items: got %d, want %d", got, want)
				}
				if got, want := tt.args.pred.Continue, `{"offset":5}`; got != want {
					t.Errorf("bad continue token: got %q, want %q", got, want)
				}
				tt.args.pred.Limit = 6

				// get the rest of the list
				list, err = s.List(ctx, tt.args.req, tt.args.pred)
				if err != nil {
					t.Fatal(err)
				}
				if got, want := list.Len(), 5; got != want {
					t.Errorf("wrong number of items: got %d, want %d", got, want)
				}
				if tt.args.pred.Continue != "" {
					t.Error("expected empty continue token")
				}
				tt.args.pred.Limit = 5

				// Test listing in descending order
				tt.args.pred.Continue = ""
				tt.args.req.SortOrder = storev2.SortDescend
				list, err = s.List(ctx, tt.args.req, tt.args.pred)
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
				tt.args.pred.Continue = ""
				tt.args.req.SortOrder = storev2.SortAscend
				list, err = s.List(ctx, tt.args.req, tt.args.pred)
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
		})
	}
}

func TestStoreExists(t *testing.T) {
	type args struct {
		req     storev2.ResourceRequest
		wrapper storev2.Wrapper
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.Interface)
	}{
		{
			name: "can check if an entity config exists",
			args: func() args {
				cfg := corev3.FixtureEntityConfig("foo")
				req := storev2.NewResourceRequestFromResource(cfg)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapEntityConfig(cfg),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
			},
		},
		{
			name: "can check if an entity state exists",
			args: func() args {
				state := corev3.FixtureEntityState("foo")
				req := storev2.NewResourceRequestFromResource(state)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapEntityState(state),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
				testCreateEntityConfig(t, s, "foo")
			},
		},
		{
			name: "can check if a namespace exists",
			args: func() args {
				ns := corev3.FixtureNamespace("bar")
				req := storev2.NewResourceRequestFromResource(ns)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapNamespace(ns),
				}
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWithPostgresStoreV2(t, func(s storev2.Interface) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, s)
				}
				ctx := context.Background()
				// Exists should return false
				got, err := s.Exists(ctx, tt.args.req)
				if err != nil {
					t.Fatal(err)
				}
				if want := false; got != want {
					t.Errorf("got true, want false")
				}

				// CreateIfNotExists should succeed
				if err := s.CreateIfNotExists(ctx, tt.args.req, tt.args.wrapper); err != nil {
					t.Fatal(err)
				}
				got, err = s.Exists(ctx, tt.args.req)
				if err != nil {
					t.Fatal(err)
				}
				if want := true; got != want {
					t.Errorf("got false, want true")
				}
			})
		})
	}
}

func TestStorePatch(t *testing.T) {
	type args struct {
		req     storev2.ResourceRequest
		wrapper storev2.Wrapper
		patcher patch.Patcher
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.Interface)
	}{
		{
			name: "an entity config can be patched",
			args: func() args {
				cfg := corev3.FixtureEntityConfig("foo")
				req := storev2.NewResourceRequestFromResource(cfg)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapEntityConfig(cfg),
					patcher: &patch.Merge{
						MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
					},
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
			},
		},
		{
			name: "an entity state can be patched",
			args: func() args {
				state := corev3.FixtureEntityState("foo")
				req := storev2.NewResourceRequestFromResource(state)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapEntityState(state),
					patcher: &patch.Merge{
						MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
					},
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
				testCreateEntityConfig(t, s, "foo")
			},
		},
		{
			name: "a namespace can be patched",
			args: func() args {
				ns := corev3.FixtureNamespace("bar")
				req := storev2.NewResourceRequestFromResource(ns)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapNamespace(ns),
					patcher: &patch.Merge{
						MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
					},
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				testCreateNamespace(t, s, "default")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWithPostgresStoreV2(t, func(s storev2.Interface) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, s)
				}
				ctx := context.Background()
				if err := s.CreateOrUpdate(ctx, tt.args.req, tt.args.wrapper); err != nil {
					t.Error(err)
				}
				if err := s.Patch(ctx, tt.args.req, tt.args.wrapper, tt.args.patcher, nil); err != nil {
					t.Fatal(err)
				}

				updatedWrapper, err := s.Get(ctx, tt.args.req)
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
		})
	}
}

func TestStoreGetMultiple(t *testing.T) {
	type args struct {
		req     storev2.ResourceRequest
		wrapper storev2.Wrapper
	}
	tests := []struct {
		name string
		args args
		reqs func(*testing.T, storev2.Interface) []storev2.ResourceRequest
		test func(*testing.T, storev2.Wrapper)
	}{
		{
			name: "multiple entity configs can be retrieved",
			args: func() args {
				cfg := corev3.FixtureEntityConfig("foo")
				req := storev2.NewResourceRequestFromResource(cfg)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapEntityConfig(cfg),
				}
			}(),
			reqs: func(t *testing.T, s storev2.Interface) []storev2.ResourceRequest {
				testCreateNamespace(t, s, "default")
				reqs := make([]storev2.ResourceRequest, 0)
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					cfg := corev3.FixtureEntityConfig(entityName)
					req := storev2.NewResourceRequestFromResource(cfg)
					req.UsePostgres = true
					reqs = append(reqs, req)
					testCreateEntityConfig(t, s, entityName)
				}
				return reqs
			},
			test: func(t *testing.T, wrapper storev2.Wrapper) {
				var cfg corev3.EntityConfig
				if err := wrapper.UnwrapInto(&cfg); err != nil {
					t.Error(err)
				}

				if got, want := len(cfg.Subscriptions), 2; got != want {
					t.Errorf("wrong number of subscriptions, got = %v, want %v", got, want)
				}
				if got, want := len(cfg.KeepaliveHandlers), 1; got != want {
					t.Errorf("wrong number of keepalive handlers, got = %v, want %v", got, want)
				}
			},
		},
		{
			name: "multiple entity states can be retrieved",
			args: func() args {
				state := corev3.FixtureEntityState("foo")
				req := storev2.NewResourceRequestFromResource(state)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapEntityState(state),
				}
			}(),
			reqs: func(t *testing.T, s storev2.Interface) []storev2.ResourceRequest {
				testCreateNamespace(t, s, "default")
				reqs := make([]storev2.ResourceRequest, 0)
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					state := corev3.FixtureEntityState(entityName)
					req := storev2.NewResourceRequestFromResource(state)
					req.UsePostgres = true
					reqs = append(reqs, req)
					testCreateEntityConfig(t, s, entityName)
					testCreateEntityState(t, s, entityName)
				}
				return reqs
			},
			test: func(t *testing.T, wrapper storev2.Wrapper) {
				var state corev3.EntityState
				if err := wrapper.UnwrapInto(&state); err != nil {
					t.Error(err)
				}

				if got, want := state.LastSeen, int64(12345); got != want {
					t.Errorf("wrong last_seen value, got = %v, want = %v", got, want)
				}
				if got, want := state.System.Arch, "amd64"; got != want {
					t.Errorf("wrong system arch value, got = %v, want %v", got, want)
				}
			},
		},
		{
			name: "multiple namespaces can be retrieved",
			args: func() args {
				ns := corev3.FixtureNamespace("bar")
				req := storev2.NewResourceRequestFromResource(ns)
				req.UsePostgres = true
				return args{
					req:     req,
					wrapper: WrapNamespace(ns),
				}
			}(),
			reqs: func(t *testing.T, s storev2.Interface) []storev2.ResourceRequest {
				reqs := make([]storev2.ResourceRequest, 0)
				for i := 0; i < 10; i++ {
					namespaceName := fmt.Sprintf("foo-%d", i)
					namespace := corev3.FixtureNamespace(namespaceName)
					req := storev2.NewResourceRequestFromResource(namespace)
					req.UsePostgres = true
					reqs = append(reqs, req)
					testCreateNamespace(t, s, namespaceName)
				}
				return reqs
			},
			test: func(t *testing.T, wrapper storev2.Wrapper) {
				var namespace corev3.Namespace
				if err := wrapper.UnwrapInto(&namespace); err != nil {
					t.Error(err)
				}

				if got, want := len(namespace.Metadata.Labels), 0; got != want {
					t.Errorf("wrong number of labels, got = %v, want = %v", got, want)
				}
				if got, want := len(namespace.Metadata.Annotations), 0; got != want {
					t.Errorf("wrong number of annotations, got = %v, want = %v", got, want)
				}
				if got, want := namespace.Metadata.Name, "foo-"; !strings.Contains(got, want) {
					t.Errorf("wrong namespace name, got = %v, want name to contain = %v", got, want)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWithPostgresStoreV2(t, func(s storev2.Interface) {
				reqs := tt.reqs(t, s)
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
					tt.test(t, wrapper)
				}
				req := reqs[0]
				req.Name = "notexists"
				result, err = s.(*StoreV2).GetMultiple(context.Background(), []storev2.ResourceRequest{req})
				if err != nil {
					t.Fatal(err)
				}
				if got, want := len(result), 0; got != want {
					t.Fatalf("wrong result length: got %d, want %d", got, want)
				}
			})
		})
	}
}
