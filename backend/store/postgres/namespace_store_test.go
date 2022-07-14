package postgres

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-test/deep"
	"github.com/jackc/pgx/v4/pgxpool"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

func TestNamespaceStore_CreateIfNotExists(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace *corev3.Namespace
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.Interface)
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "bar")
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewStoreV2(db))
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
		beforeHook func(*testing.T, storev2.Interface)
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
			//verifyQuery: fmt.Sprintf("SELECT * FROM %s", namespaceStoreName),
			//want:        1,
		},
		{
			name: "updates when namespace exists",
			args: func() args {
				return args{
					ctx:       context.Background(),
					namespace: corev3.FixtureNamespace("foo"),
				}
			}(),
			//verifyQuery: fmt.Sprintf("SELECT * FROM %s", namespaceStoreName),
			//want:        1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewStoreV2(db))
				}
				s := &NamespaceStore{
					db: db,
				}
				if err := s.CreateOrUpdate(tt.args.ctx, tt.args.namespace); (err != nil) != tt.wantErr {
					t.Errorf("NamespaceStore.CreateOrUpdate() error = %v, wantErr %v", err, tt.wantErr)
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
		beforeHook func(*testing.T, storev2.Interface)
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewStoreV2(db))
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
		beforeHook func(*testing.T, storev2.Interface)
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "bar")
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewStoreV2(db))
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
		beforeHook func(*testing.T, storev2.Interface)
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "foo")
			},
			want: corev3.NewNamespace("foo"),
			// ns := corev3.FixtureNamespace("foo")
			// wrapper := WrapNamespace(ns).(*NamespaceWrapper)
			// wrapper.ID = 1
			// wrapper.CreatedAt = time.Now()
			// wrapper.UpdatedAt = time.Now()
			// return wrapper
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewStoreV2(db))
				}
				s := &NamespaceStore{
					db: db,
				}
				got, err := s.Get(tt.args.ctx, tt.args.name)
				if (err != nil) != tt.wantErr {
					t.Errorf("NamespaceStore.Get() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NamespaceStore.Get() = %v, want %v", got, tt.want)
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
		beforeHook func(*testing.T, storev2.Interface)
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewStoreV2(db))
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
		beforeHook func(*testing.T, storev2.Interface)
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "bar")
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewStoreV2(db))
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
		beforeHook func(*testing.T, storev2.Interface)
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
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
					tt.beforeHook(t, NewStoreV2(db))
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
					t.Errorf("NamespaceStore.List() mismatch: %v", diff)
				}
				// if !reflect.DeepEqual(got, tt.want) {
				// 	t.Errorf("NamespaceStore.List() = %v, want %v", got, tt.want)
				// }
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
		beforeHook func(*testing.T, storev2.Interface)
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewStoreV2(db))
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
		beforeHook func(*testing.T, storev2.Interface)
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewStoreV2(db))
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
		beforeHook func(*testing.T, storev2.Interface)
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
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
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "foo")
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewStoreV2(db))
				}
				s := &NamespaceStore{
					db: db,
				}
				got, err := s.isEmpty(tt.args.ctx, tt.args.name)
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

// {
// 			name: "multiple namespaces can be retrieved",
// 			args: func() args {
// 				ns := corev3.FixtureNamespace("bar")
// 				req := storev2.NewResourceRequestFromResource(ns)
// 				return args{
// 					req:     req,
// 					wrapper: WrapNamespace(ns),
// 				}
// 			}(),
// 			reqs: func(t *testing.T, s storev2.Interface) []storev2.ResourceRequest {
// 				reqs := make([]storev2.ResourceRequest, 0)
// 				for i := 0; i < 10; i++ {
// 					namespaceName := fmt.Sprintf("foo-%d", i)
// 					namespace := corev3.FixtureNamespace(namespaceName)
// 					req := storev2.NewResourceRequestFromResource(namespace)
// 					reqs = append(reqs, req)
// 					createNamespace(t, s, namespaceName)
// 				}
// 				return reqs
// 			},
// 			test: func(t *testing.T, wrapper storev2.Wrapper) {
// 				var namespace corev3.Namespace
// 				if err := wrapper.UnwrapInto(&namespace); err != nil {
// 					t.Error(err)
// 				}

// 				if got, want := len(namespace.Metadata.Labels), 0; got != want {
// 					t.Errorf("wrong number of labels, got = %v, want = %v", got, want)
// 				}
// 				if got, want := len(namespace.Metadata.Annotations), 0; got != want {
// 					t.Errorf("wrong number of annotations, got = %v, want = %v", got, want)
// 				}
// 				if got, want := namespace.Metadata.Name, "foo-"; !strings.Contains(got, want) {
// 					t.Errorf("wrong namespace name, got = %v, want name to contain = %v", got, want)
// 				}
// 			},
// 		},
