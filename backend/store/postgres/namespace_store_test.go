package postgres

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-test/deep"
	"github.com/jackc/pgx/v5/pgxpool"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/stretchr/testify/require"
)

func TestNamespaceStore_CreateIfNotExists(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace *corev3.Namespace
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		wantErr    bool
	}{
		{
			name: "succeeds when namespace does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: corev3.FixtureNamespace("bar"),
			},
		},
		{
			name: "succeeds when namespace is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: corev3.FixtureNamespace("bar"),
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "bar")
				deleteNamespace(t, s, "bar")
			},
		},
		{
			name: "fails when namespace exists",
			args: args{
				ctx:       context.Background(),
				namespace: corev3.FixtureNamespace("bar"),
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "bar")
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &NamespaceStore{
					db: db,
				}
				if err := s.CreateIfNotExists(tt.args.ctx, tt.args.namespace); (err != nil) != tt.wantErr {
					t.Errorf("NamespaceStore.CreateIfNotExists() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		})
	}
}

func TestNamespaceStore_CreateOrUpdate(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace *corev3.Namespace
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		afterHook  func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		wantErr    bool
	}{
		{
			name: "creates when namespace does not exist",
			args: func() args {
				return args{
					ctx:       context.Background(),
					namespace: corev3.FixtureNamespace("foo"),
				}
			}(),
			afterHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				ctx := context.Background()
				namespace, err := s.Get(ctx, "foo")
				require.NoError(t, err)
				require.Equal(t, "foo", namespace.Metadata.Name)
			},
		},
		{
			name: "updates when namespace exists",
			args: func() args {
				return args{
					ctx: context.Background(),
					namespace: func() *corev3.Namespace {
						namespace := corev3.FixtureNamespace("foo")
						namespace.Metadata.Annotations["updated"] = "true"
						return namespace
					}(),
				}
			}(),
			afterHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				ctx := context.Background()
				namespace, err := s.Get(ctx, "foo")
				require.NoError(t, err)
				require.Equal(t, "foo", namespace.Metadata.Name)
				require.Equal(t, "true", namespace.Metadata.Annotations["updated"])
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &NamespaceStore{
					db: db,
				}
				if err := s.CreateOrUpdate(tt.args.ctx, tt.args.namespace); (err != nil) != tt.wantErr {
					t.Errorf("NamespaceStore.CreateOrUpdate() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.afterHook != nil {
					tt.afterHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
			})
		})
	}
}

func TestNamespaceStore_Delete(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		wantErr    bool
	}{
		{
			name: "fails when namespace does not exist",
			args: args{
				ctx:  context.Background(),
				name: "bar",
			},
			wantErr: true,
		},
		{
			name: "succeeds when namespace exists",
			args: args{
				ctx:  context.Background(),
				name: "bar",
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &NamespaceStore{
					db: db,
				}
				if err := s.Delete(tt.args.ctx, tt.args.name); (err != nil) != tt.wantErr {
					t.Errorf("NamespaceStore.Delete() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		})
	}
}

func TestNamespaceStore_Exists(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		want       bool
		wantErr    bool
	}{
		{
			name: "returns false when namespace does not exist",
			args: args{
				ctx:  context.Background(),
				name: "bar",
			},
		},
		{
			name: "returns false when namespace is soft deleted",
			args: args{
				ctx:  context.Background(),
				name: "bar",
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "bar")
				deleteNamespace(t, s, "bar")
			},
		},
		{
			name: "returns true when namespace exists",
			args: args{
				ctx:  context.Background(),
				name: "bar",
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "bar")
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &NamespaceStore{
					db: db,
				}
				got, err := s.Exists(tt.args.ctx, tt.args.name)
				if (err != nil) != tt.wantErr {
					t.Errorf("NamespaceStore.Exists() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("NamespaceStore.Exists() = %v, want %v", got, tt.want)
				}
			})
		})
	}
}

func TestNamespaceStore_Get(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		want       *corev3.Namespace
		wantErr    bool
	}{
		{
			name: "fails when namespace does not exist",
			args: args{
				ctx:  context.Background(),
				name: "foo",
			},
			wantErr: true,
		},
		{
			name: "fails when namespace is soft deleted",
			args: args{
				ctx:  context.Background(),
				name: "foo",
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "foo")
				deleteNamespace(t, s, "foo")
			},
			wantErr: true,
		},
		{
			name: "succeeds when namespace exists",
			args: args{
				ctx:  context.Background(),
				name: "foo",
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "foo")
			},
			want: func() *corev3.Namespace {
				namespace := corev3.NewNamespace("foo")
				// TODO: uncomment after ID, CreatedAt, UpdatedAt & DeletedAt
				// are added to corev3.Namespace.
				//
				// namespace.ID = 1
				// namespace.CreatedAt = time.Now()
				// namespace.UpdatedAt = time.Now()
				return namespace
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &NamespaceStore{
					db: db,
				}
				got, err := s.Get(tt.args.ctx, tt.args.name)
				if (err != nil) != tt.wantErr {
					t.Errorf("NamespaceStore.Get() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				// TODO: uncomment after ID, CreatedAt, UpdatedAt & DeletedAt
				// are added to corev3.Namespace.
				//
				// createdAtDelta := time.Since(tt.want.CreatedAt) / 2
				// wantCreatedAt := time.Now().Add(-createdAtDelta)
				// require.WithinDuration(t, wantCreatedAt, got.CreatedAt, createdAtDelta)
				// got.CreatedAt = tt.want.CreatedAt

				// updatedAtDelta := time.Since(tt.want.UpdatedAt) / 2
				// wantUpdatedAt := time.Now().Add(-updatedAtDelta)
				// require.WithinDuration(t, wantUpdatedAt, got.UpdatedAt, updatedAtDelta)
				// got.UpdatedAt = tt.want.UpdatedAt

				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NamespaceStore.Get() = %v, want %v", got, tt.want)
				}
			})
		})
	}
}

func TestNamespaceStore_GetMultiple(t *testing.T) {
	type args struct {
		ctx       context.Context
		resources []string
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		want       uniqueNamespaces
		wantErr    bool
	}{
		{
			name: "succeeds when no requested namespaces exist",
			args: args{
				ctx:       context.Background(),
				resources: []string{"default", "ops"},
			},
			want: uniqueNamespaces{},
		},
		{
			name: "succeeds when all namespaces are soft deleted",
			args: args{
				ctx:       context.Background(),
				resources: []string{"default", "ops"},
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "default")
				createNamespace(t, s, "ops")
				deleteNamespace(t, s, "default")
				deleteNamespace(t, s, "ops")
			},
			want: uniqueNamespaces{},
		},
		{
			name: "succeeds when some namespaces are soft deleted",
			args: args{
				ctx:       context.Background(),
				resources: []string{"default", "ops"},
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "default")
				createNamespace(t, s, "ops")
				deleteNamespace(t, s, "default")
			},
			want: uniqueNamespaces{
				uniqueResource{Name: "ops"}: corev3.FixtureNamespace("ops"),
			},
		},
		{
			name: "succeeds when all entity namespaces exists",
			args: args{
				ctx:       context.Background(),
				resources: []string{"default", "ops"},
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "default")
				createNamespace(t, s, "ops")
			},
			want: uniqueNamespaces{
				uniqueResource{Name: "default"}: corev3.FixtureNamespace("default"),
				uniqueResource{Name: "ops"}:     corev3.FixtureNamespace("ops"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &NamespaceStore{
					db: db,
				}
				got, err := s.GetMultiple(tt.args.ctx, tt.args.resources)
				if (err != nil) != tt.wantErr {
					t.Errorf("NamespaceStore.GetMultiple() error = %v, wantErr %v", err, tt.wantErr)
				}
				if diff := deep.Equal(got, tt.want); diff != nil {
					t.Errorf("NamespaceStore.GetMultiple() got differs from want: %v", diff)
				}
			})
		})
	}
}

func TestNamespaceStore_HardDelete(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		wantErr    bool
	}{
		{
			name: "fails when namespace does not exist",
			args: args{
				ctx:  context.Background(),
				name: "bar",
			},
			wantErr: true,
		},
		{
			name: "succeeds when namespace is soft deleted",
			args: args{
				ctx:  context.Background(),
				name: "bar",
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "bar")
				deleteNamespace(t, s, "bar")
			},
		},
		{
			name: "succeeds when namespace exists",
			args: args{
				ctx:  context.Background(),
				name: "bar",
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &NamespaceStore{
					db: db,
				}
				if err := s.HardDelete(tt.args.ctx, tt.args.name); (err != nil) != tt.wantErr {
					t.Errorf("NamespaceStore.HardDelete() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		})
	}
}

func TestNamespaceStore_HardDeleted(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		want       bool
		wantErr    bool
	}{
		{
			name: "returns true when namespace does not exist",
			args: args{
				ctx:  context.Background(),
				name: "bar",
			},
			want: true,
		},
		{
			name: "returns false when namespace is soft deleted",
			args: args{
				ctx:  context.Background(),
				name: "bar",
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "bar")
				deleteNamespace(t, s, "bar")
			},
			want: false,
		},
		{
			name: "returns false when namespace exists",
			args: args{
				ctx:  context.Background(),
				name: "bar",
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "bar")
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &NamespaceStore{
					db: db,
				}
				got, err := s.HardDeleted(tt.args.ctx, tt.args.name)
				if (err != nil) != tt.wantErr {
					t.Errorf("NamespaceStore.HardDeleted() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("NamespaceStore.HardDeleted() = %v, want %v", got, tt.want)
				}
			})
		})
	}
}

func TestNamespaceStore_List(t *testing.T) {
	type args struct {
		ctx  context.Context
		pred *store.SelectionPredicate
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		want       []*corev3.Namespace
		wantErr    bool
	}{
		{
			name: "succeeds when no namespaces exist",
			args: args{
				ctx: context.Background(),
			},
			want: []*corev3.Namespace{},
		},
		{
			name: "succeeds when namespaces exist",
			args: args{
				ctx: context.Background(),
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				for i := 0; i < 10; i++ {
					namespaceName := fmt.Sprintf("foo-%d", i)
					createNamespace(t, s, namespaceName)
				}
			},
			want: func() []*corev3.Namespace {
				namespaces := []*corev3.Namespace{}
				for i := 0; i < 10; i++ {
					namespaceName := fmt.Sprintf("foo-%d", i)
					namespace := corev3.FixtureNamespace(namespaceName)
					namespaces = append(namespaces, namespace)
				}
				return namespaces
			}(),
		},
		{
			name: "succeeds when limit set and namespaces exist",
			args: args{
				ctx:  context.Background(),
				pred: &store.SelectionPredicate{Limit: 5},
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				for i := 0; i < 10; i++ {
					namespaceName := fmt.Sprintf("foo-%d", i)
					createNamespace(t, s, namespaceName)
				}
			},
			want: []*corev3.Namespace{
				corev3.FixtureNamespace("foo-0"),
				corev3.FixtureNamespace("foo-1"),
				corev3.FixtureNamespace("foo-2"),
				corev3.FixtureNamespace("foo-3"),
				corev3.FixtureNamespace("foo-4"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &NamespaceStore{
					db: db,
				}
				got, err := s.List(tt.args.ctx, tt.args.pred)
				if (err != nil) != tt.wantErr {
					t.Errorf("NamespaceStore.List() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if diff := deep.Equal(got, tt.want); len(diff) > 0 {
					t.Errorf("NamespaceStore.List() got differs from want: %v", diff)
				}
				// if !reflect.DeepEqual(got, tt.want) {
				// 	t.Errorf("NamespaceStore.List() = %v, want %v", got, tt.want)
				// }
			})
		})
	}
}

func TestNamespaceStore_Count(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		want       int
		wantErr    bool
	}{
		{
			name: "succeeds when no namespaces exist",
			args: args{
				ctx: context.Background(),
			},
		},
		{
			name: "succeeds when namespaces exist",
			args: args{
				ctx: context.Background(),
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				for i := 0; i < 10; i++ {
					namespaceName := fmt.Sprintf("foo-%d", i)
					createNamespace(t, s, namespaceName)
				}
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &NamespaceStore{
					db: db,
				}
				got, err := s.Count(tt.args.ctx)
				if (err != nil) != tt.wantErr {
					t.Errorf("NamespaceStore.Count() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if diff := deep.Equal(got, tt.want); len(diff) > 0 {
					t.Errorf("NamespaceStore.Count() got differs from want: %v", diff)
				}
			})
		})
	}
}
func TestNamespaceStore_Patch(t *testing.T) {
	type args struct {
		ctx        context.Context
		name       string
		patcher    patch.Patcher
		conditions *store.ETagCondition
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		wantErr    bool
	}{
		{
			name: "fails when namespace does not exist",
			args: args{
				ctx:  context.Background(),
				name: "bar",
				patcher: &patch.Merge{
					MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
				},
			},
			wantErr: true,
		},
		{
			name: "fails when namespace is soft deleted",
			args: args{
				ctx:  context.Background(),
				name: "bar",
				patcher: &patch.Merge{
					MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
				},
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "bar")
				deleteNamespace(t, s, "bar")
			},
			wantErr: true,
		},
		{
			name: "succeeds when namespace exists",
			args: args{
				ctx:  context.Background(),
				name: "bar",
				patcher: &patch.Merge{
					MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
				},
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &NamespaceStore{
					db: db,
				}
				if err := s.Patch(tt.args.ctx, tt.args.name, tt.args.patcher, tt.args.conditions); (err != nil) != tt.wantErr {
					t.Errorf("NamespaceStore.Patch() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		})
	}
}

func TestNamespaceStore_UpdateIfExists(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace *corev3.Namespace
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		wantErr    bool
	}{
		{
			name: "fails when namespace does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: corev3.FixtureNamespace("bar"),
			},
			wantErr: true,
		},
		{
			name: "succeeds when namespace is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: corev3.FixtureNamespace("bar"),
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "bar")
				deleteNamespace(t, s, "bar")
			},
		},
		{
			name: "succeeds when namespace exists",
			args: args{
				ctx:       context.Background(),
				namespace: corev3.FixtureNamespace("bar"),
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &NamespaceStore{
					db: db,
				}
				if err := s.UpdateIfExists(tt.args.ctx, tt.args.namespace); (err != nil) != tt.wantErr {
					t.Errorf("NamespaceStore.UpdateIfExists() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		})
	}
}

func TestNamespaceStore_isEmpty(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		want       bool
		wantErr    bool
	}{
		{
			name: "fails when namespace does not exist",
			args: args{
				ctx:  context.Background(),
				name: "foo",
			},
			wantErr: true,
		},
		{
			name: "returns true when namespace is empty",
			args: args{
				ctx:  context.Background(),
				name: "default",
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "default")
			},
			want: true,
		},
		{
			name: "returns false when namespace contains resources",
			args: args{
				ctx:  context.Background(),
				name: "default",
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, ec storev2.EntityConfigStore) {
				createNamespace(t, s, "default")
				createEntityConfig(t, ec, "default", "foo")
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &NamespaceStore{
					db: db,
				}
				got, err := s.IsEmpty(tt.args.ctx, tt.args.name)
				if (err != nil) != tt.wantErr {
					t.Errorf("NamespaceStore.isEmpty() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("NamespaceStore.isEmpty() = %v, want %v", got, tt.want)
				}
			})
		})
	}
}
