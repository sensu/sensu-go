package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/stretchr/testify/require"
)

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
				return args{
					req:     req,
					wrapper: WrapEntityConfig(cfg),
				}
			}(),
			verifyQuery: fmt.Sprintf("SELECT * FROM %s", entityConfigStoreName),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
			want: 1,
		},
		{
			name: "entity states can be created and updated",
			args: func() args {
				state := corev3.FixtureEntityState("foo")
				req := storev2.NewResourceRequestFromResource(state)
				return args{
					req:     req,
					wrapper: WrapEntityState(state),
				}
			}(),
			verifyQuery: fmt.Sprintf("SELECT * FROM %s", entityStateStoreName),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "foo")
			},
			want: 1,
		},
		{
			name: "namespaces can be created and updated",
			args: func() args {
				ns := corev3.FixtureNamespace("foo")
				req := storev2.NewResourceRequestFromResource(ns)
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
				require.NoError(t, err)
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
				return args{
					req:     req,
					wrapper: WrapEntityConfig(cfg),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
		},
		{
			name: "entity states can be updated if one exists",
			args: func() args {
				state := corev3.FixtureEntityState("foo")
				req := storev2.NewResourceRequestFromResource(state)
				return args{
					req:     req,
					wrapper: WrapEntityState(state),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "foo")
			},
		},
		{
			name: "namespaces can be updated if one exists",
			args: func() args {
				ns := corev3.FixtureNamespace("bar")
				req := storev2.NewResourceRequestFromResource(ns)
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
					t.Fatal("expected non-nil error")
				} else {
					var e *store.ErrNotFound
					if !errors.As(err, &e) {
						t.Fatalf("expected %T, got %T: %s", e, err, err)
					}
				}
				if err := s.CreateOrUpdate(ctx, tt.args.req, tt.args.wrapper); err != nil {
					t.Fatal(err)
				}

				// UpdateIfExists should succeed
				if err := s.UpdateIfExists(ctx, tt.args.req, tt.args.wrapper); err != nil {
					t.Fatal(err)
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
				return args{
					req:     req,
					wrapper: WrapEntityConfig(cfg),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
		},
		{
			name: "entity states can be created if one does not exist",
			args: func() args {
				state := corev3.FixtureEntityState("foo")
				req := storev2.NewResourceRequestFromResource(state)
				return args{
					req:     req,
					wrapper: WrapEntityState(state),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "foo")
			},
		},
		{
			name: "namespaces can be created if one does not exist",
			args: func() args {
				ns := corev3.FixtureNamespace("bar")
				req := storev2.NewResourceRequestFromResource(ns)
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
					t.Fatal("expected non-nil error")
				} else {
					var e *store.ErrAlreadyExists
					if !errors.As(err, &e) {
						t.Fatalf("expected %T, got %T: %s", e, err, err)
					}
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
		want       wrapperWithStatus
	}{
		{
			name: "an entity config can be retrieved",
			args: func() args {
				cfg := corev3.FixtureEntityConfig("foo")
				req := storev2.NewResourceRequestFromResource(cfg)
				return args{
					req: req,
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "foo")
			},
			want: func() wrapperWithStatus {
				cfg := corev3.FixtureEntityConfig("foo")
				wrapper := WrapEntityConfig(cfg).(*EntityConfigWrapper)
				wrapper.ID = 1
				wrapper.NamespaceID = 1
				wrapper.CreatedAt = time.Now()
				wrapper.UpdatedAt = time.Now()
				return wrapper
			}(),
		},
		{
			name: "an entity state can be retrieved",
			args: func() args {
				state := corev3.FixtureEntityState("foo")
				req := storev2.NewResourceRequestFromResource(state)
				return args{
					req: req,
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "foo")
				createEntityState(t, s, "foo")
			},
			want: func() wrapperWithStatus {
				state := corev3.FixtureEntityState("foo")
				wrapper := WrapEntityState(state).(*EntityStateWrapper)
				wrapper.ID = 1
				wrapper.NamespaceID = 1
				wrapper.EntityConfigID = 1
				wrapper.CreatedAt = time.Now()
				wrapper.UpdatedAt = time.Now()
				return wrapper
			}(),
		},
		{
			name: "a namespace can be retrieved",
			args: func() args {
				ns := corev3.FixtureNamespace("foo")
				req := storev2.NewResourceRequestFromResource(ns)
				return args{
					req: req,
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "foo")
			},
			want: func() wrapperWithStatus {
				ns := corev3.FixtureNamespace("foo")
				wrapper := WrapNamespace(ns).(*NamespaceWrapper)
				wrapper.ID = 1
				wrapper.CreatedAt = time.Now()
				wrapper.UpdatedAt = time.Now()
				return wrapper
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWithPostgresStoreV2(t, func(s storev2.Interface) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, s)
				}

				wrapper, err := s.Get(context.Background(), tt.args.req)
				if err != nil {
					t.Fatal(err)
				}
				got := wrapper.(wrapperWithStatus)

				createdAtDelta := time.Since(tt.want.GetCreatedAt()) / 2
				wantCreatedAt := time.Now().Add(-createdAtDelta)
				require.WithinDuration(t, wantCreatedAt, got.GetCreatedAt(), createdAtDelta)
				got.SetCreatedAt(tt.want.GetCreatedAt())

				updatedAtDelta := time.Since(tt.want.GetUpdatedAt()) / 2
				wantUpdatedAt := time.Now().Add(-updatedAtDelta)
				require.WithinDuration(t, wantUpdatedAt, got.GetUpdatedAt(), updatedAtDelta)
				got.SetUpdatedAt(tt.want.GetUpdatedAt())

				if diff := deep.Equal(got, tt.want); diff != nil {
					t.Fatal(diff)
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
			name: "an entity config can be soft deleted",
			args: func() args {
				cfg := corev3.FixtureEntityConfig("foo")
				req := storev2.NewResourceRequestFromResource(cfg)
				return args{
					req:     req,
					wrapper: WrapEntityConfig(cfg),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
		},
		{
			name: "an entity state can be soft deleted",
			args: func() args {
				state := corev3.FixtureEntityState("foo")
				req := storev2.NewResourceRequestFromResource(state)
				return args{
					req:     req,
					wrapper: WrapEntityState(state),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "foo")
			},
		},
		{
			name: "a namespace can be soft deleted",
			args: func() args {
				ns := corev3.FixtureNamespace("bar")
				req := storev2.NewResourceRequestFromResource(ns)
				return args{
					req:     req,
					wrapper: WrapNamespace(ns),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
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
				// Resource should no longer exist
				exists, err := s.(*StoreV2).Exists(ctx, tt.args.req)
				if err != nil {
					t.Fatal(err)
				}
				require.False(t, exists)
			})
		})
	}
}

func TestStoreHardDelete(t *testing.T) {
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
			name: "an entity config can be hard deleted",
			args: func() args {
				cfg := corev3.FixtureEntityConfig("foo")
				req := storev2.NewResourceRequestFromResource(cfg)
				return args{
					req:     req,
					wrapper: WrapEntityConfig(cfg),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
		},
		{
			name: "an entity state can be hard deleted",
			args: func() args {
				state := corev3.FixtureEntityState("foo")
				req := storev2.NewResourceRequestFromResource(state)
				return args{
					req:     req,
					wrapper: WrapEntityState(state),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "foo")
			},
		},
		{
			name: "a namespace can be hard deleted",
			args: func() args {
				ns := corev3.FixtureNamespace("bar")
				req := storev2.NewResourceRequestFromResource(ns)
				return args{
					req:     req,
					wrapper: WrapNamespace(ns),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWithPostgresStoreV2(t, func(s storev2.Interface) {
				ctx := context.Background()
				if tt.beforeHook != nil {
					tt.beforeHook(t, s)
				}

				// CreateIfNotExists should succeed
				if err := s.CreateIfNotExists(ctx, tt.args.req, tt.args.wrapper); err != nil {
					t.Fatal(err)
				}
				if err := s.(*StoreV2).HardDelete(ctx, tt.args.req); err != nil {
					t.Fatal(err)
				}
				exists, err := s.(*StoreV2).Exists(ctx, tt.args.req)
				if err != nil {
					t.Fatal(err)
				}
				require.False(t, exists)
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
				typeMeta := corev2.TypeMeta{Type: "EntityConfig", APIVersion: "core/v3"}
				req := storev2.NewResourceRequest(typeMeta, "default", "anything", entityConfigStoreName)
				pred := &store.SelectionPredicate{Limit: 5}
				return args{
					req:  req,
					pred: pred,
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					createEntityConfig(t, s, entityName)
				}
			},
		},
		{
			name: "entity states can be listed",
			args: func() args {
				typeMeta := corev2.TypeMeta{Type: "EntityState", APIVersion: "core/v3"}
				req := storev2.NewResourceRequest(typeMeta, "default", "anything", entityStateStoreName)
				pred := &store.SelectionPredicate{Limit: 5}
				return args{
					req:  req,
					pred: pred,
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					createEntityConfig(t, s, entityName)
					createEntityState(t, s, entityName)
				}
			},
		},
		{
			name: "namespaces can be listed",
			args: func() args {
				typeMeta := corev2.TypeMeta{Type: "Namespace", APIVersion: "core/v3"}
				req := storev2.NewResourceRequest(typeMeta, "", "anything", namespaceStoreName)
				pred := &store.SelectionPredicate{Limit: 5}
				return args{
					req:  req,
					pred: pred,
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				for i := 0; i < 10; i++ {
					namespaceName := fmt.Sprintf("foo-%d", i)
					createNamespace(t, s, namespaceName)
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
				return args{
					req:     req,
					wrapper: WrapEntityConfig(cfg),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
		},
		{
			name: "can check if an entity state exists",
			args: func() args {
				state := corev3.FixtureEntityState("foo")
				req := storev2.NewResourceRequestFromResource(state)
				return args{
					req:     req,
					wrapper: WrapEntityState(state),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "foo")
			},
		},
		{
			name: "can check if a namespace exists",
			args: func() args {
				ns := corev3.FixtureNamespace("bar")
				req := storev2.NewResourceRequestFromResource(ns)
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
				return args{
					req:     req,
					wrapper: WrapEntityConfig(cfg),
					patcher: &patch.Merge{
						MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
					},
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
		},
		{
			name: "an entity state can be patched",
			args: func() args {
				state := corev3.FixtureEntityState("foo")
				req := storev2.NewResourceRequestFromResource(state)
				return args{
					req:     req,
					wrapper: WrapEntityState(state),
					patcher: &patch.Merge{
						MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
					},
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "foo")
			},
		},
		{
			name: "a namespace can be patched",
			args: func() args {
				ns := corev3.FixtureNamespace("bar")
				req := storev2.NewResourceRequestFromResource(ns)
				return args{
					req:     req,
					wrapper: WrapNamespace(ns),
					patcher: &patch.Merge{
						MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
					},
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
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
				return args{
					req:     req,
					wrapper: WrapEntityConfig(cfg),
				}
			}(),
			reqs: func(t *testing.T, s storev2.Interface) []storev2.ResourceRequest {
				createNamespace(t, s, "default")
				reqs := make([]storev2.ResourceRequest, 0)
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					cfg := corev3.FixtureEntityConfig(entityName)
					req := storev2.NewResourceRequestFromResource(cfg)
					reqs = append(reqs, req)
					createEntityConfig(t, s, entityName)
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
				return args{
					req:     req,
					wrapper: WrapEntityState(state),
				}
			}(),
			reqs: func(t *testing.T, s storev2.Interface) []storev2.ResourceRequest {
				createNamespace(t, s, "default")
				reqs := make([]storev2.ResourceRequest, 0)
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					state := corev3.FixtureEntityState(entityName)
					req := storev2.NewResourceRequestFromResource(state)
					reqs = append(reqs, req)
					createEntityConfig(t, s, entityName)
					createEntityState(t, s, entityName)
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
					reqs = append(reqs, req)
					createNamespace(t, s, namespaceName)
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

func TestWatchEntityConfig(t *testing.T) {
	testWithPostgresStoreV2(t, func(s storev2.Interface) {
		stor, ok := s.(*StoreV2)
		if !ok {
			t.Fatal("expected storev2")
		}
		stor.watchInterval = time.Millisecond * 10
		stor.watchTxnWindow = time.Second

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		typeMeta := corev2.TypeMeta{Type: "EntityConfig", APIVersion: "core/v3"}
		watchChannel := stor.Watch(ctx, storev2.NewResourceRequest(typeMeta, "", "", entityConfigStoreName))

		select {
		case record, ok := <-watchChannel:
			t.Errorf("expected watch channel to be empty. Got %v, %v", record, ok)
		default:
			// OK
		}

		createNamespace(t, s, "default")
		createEntityConfig(t, s, "foo")

		var entityConfig corev3.EntityConfig
		select {
		case watchEvents, ok := <-watchChannel:
			if !ok {
				t.Errorf("unexpected watcher close")
			}
			event := watchEvents[0]
			if watchErr := event.Err; watchErr != nil {
				t.Errorf("unexpected watcher error %v", watchErr)
			}
			if event.Key.Name != "foo" || event.Key.Namespace != "default" {
				t.Errorf("expected name 'foo' namespace 'default', got %v, %v", event.Key.Name, event.Key.Namespace)
			}
			if event.Type != storev2.WatchCreate {
				t.Errorf("expected event type (%v), got %v", storev2.WatchCreate, event.Type)
			}
			if werr := event.Value.UnwrapInto(&entityConfig); werr != nil {
				t.Fatal(werr)
			}
		case <-time.After(time.Millisecond * 100):
			t.Fatalf("expected entity change notification but timed out")
		}

		delReq := storev2.NewResourceRequestFromResource(&entityConfig)
		if err := s.Delete(ctx, delReq); err != nil {
			t.Fatalf("could not delete entity config: %v", err)
		}

		select {
		case watchEvents, ok := <-watchChannel:
			if !ok {
				t.Errorf("unexpected watcher close")
			}
			event := watchEvents[0]
			if event.Key.Name != "foo" || event.Key.Namespace != "default" {
				t.Errorf("expected name 'foo' namespace 'default', got %v, %v", event.Key.Name, event.Key.Namespace)
			}
			if event.Type != storev2.WatchDelete {
				t.Errorf("expected event type (%v), got %v", storev2.WatchDelete, event.Type)
			}
			if werr := event.Value.UnwrapInto(&entityConfig); werr != nil {
				t.Fatal(werr)
			}
		case <-time.After(time.Millisecond * 100):
			t.Errorf("expected entity delete notification but timed out")
		}
	})
}

func TestInitialize(t *testing.T) {
	testWithPostgresStoreV2(t, func(s storev2.Interface) {
		ctx := context.Background()

		iErr := s.Initialize(ctx, func(ctx context.Context) error {
			namespace := corev3.NewNamespace("foo")
			wrapper, err := storev2.WrapResource(namespace)
			require.NoError(t, err)

			req := storev2.NewResourceRequestFromResource(namespace)

			return s.CreateIfNotExists(ctx, req, wrapper)
		})
		require.NoError(t, iErr)

		var namespace corev3.Namespace
		req := storev2.NewResourceRequestFromResource(corev3.NewNamespace("foo"))

		wrapper, err := s.Get(ctx, req)
		require.NoError(t, err)

		wErr := wrapper.UnwrapInto(&namespace)
		require.NoError(t, wErr)
		require.Equal(t, "foo", namespace.Metadata.Name)
	})
}
