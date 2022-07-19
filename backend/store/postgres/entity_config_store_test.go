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
	"github.com/stretchr/testify/require"
)

func TestEntityConfigStore_CreateIfNotExists(t *testing.T) {
	type args struct {
		ctx          context.Context
		entityConfig *corev3.EntityConfig
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
				ctx:          context.Background(),
				entityConfig: corev3.FixtureEntityConfig("bar"),
			},
		},
		{
			name: "fails when namespace is soft deleted",
			args: args{
				ctx:          context.Background(),
				entityConfig: corev3.FixtureEntityConfig("bar"),
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				deleteNamespace(t, s, "default")
			},
		},
		{
			name: "succeeds when entity config does not exist",
			args: args{
				ctx:          context.Background(),
				entityConfig: corev3.FixtureEntityConfig("bar"),
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
		},
		{
			name: "succeeds when entity config is soft deleted",
			args: args{
				ctx:          context.Background(),
				entityConfig: corev3.FixtureEntityConfig("bar"),
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "default", "bar")
				deleteEntityConfig(t, s, "default", "bar")
			},
		},
		{
			name: "fails when entity config exists",
			args: args{
				ctx:          context.Background(),
				entityConfig: corev3.FixtureEntityConfig("bar"),
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "default", "bar")
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
				s := &EntityConfigStore{
					db: db,
				}
				if err := s.CreateIfNotExists(tt.args.ctx, tt.args.entityConfig); (err != nil) != tt.wantErr {
					t.Errorf("EntityConfigStore.CreateIfNotExists() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		})
	}
}

func TestEntityConfigStore_CreateOrUpdate(t *testing.T) {
	type args struct {
		ctx          context.Context
		entityConfig *corev3.EntityConfig
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.Interface)
		afterHook  func(*testing.T, storev2.Interface)
		wantErr    bool
	}{
		{
			name: "fails when namespace does not exist",
			args: func() args {
				return args{
					ctx:          context.Background(),
					entityConfig: corev3.FixtureEntityConfig("foo"),
				}
			}(),
		},
		{
			name: "fails when namespace is soft deleted",
			args: func() args {
				return args{
					ctx:          context.Background(),
					entityConfig: corev3.FixtureEntityConfig("foo"),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				deleteNamespace(t, s, "default")
			},
		},
		{
			name: "creates when entity config does not exist",
			args: func() args {
				return args{
					ctx:          context.Background(),
					entityConfig: corev3.FixtureEntityConfig("foo"),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
			afterHook: func(t *testing.T, s storev2.Interface) {
				ctx := context.Background()
				config, err := s.EntityConfigStore().Get(ctx, "default", "foo")
				require.NoError(t, err)
				require.Equal(t, "foo", config.Metadata.Name)
			},
		},
		{
			name: "updates when entity config exists",
			args: func() args {
				return args{
					ctx: context.Background(),
					entityConfig: func() *corev3.EntityConfig {
						config := corev3.FixtureEntityConfig("foo")
						config.Metadata.Annotations["updated"] = "true"
						return config
					}(),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "default", "foo")
			},
			afterHook: func(t *testing.T, s storev2.Interface) {
				ctx := context.Background()
				config, err := s.EntityConfigStore().Get(ctx, "default", "foo")
				require.NoError(t, err)
				require.Equal(t, "foo", config.Metadata.Name)
				require.Equal(t, "true", config.Metadata.Annotations["updated"])
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				stor := NewStoreV2(db)
				if tt.beforeHook != nil {
					tt.beforeHook(t, stor)
				}
				s := &EntityConfigStore{
					db: db,
				}
				if err := s.CreateOrUpdate(tt.args.ctx, tt.args.entityConfig); (err != nil) != tt.wantErr {
					t.Errorf("EntityConfigStore.CreateOrUpdate() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.afterHook != nil {
					tt.afterHook(t, stor)
				}
			})
		})
	}
}

func TestEntityConfigStore_Delete(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace string
		name      string
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
				namespace: "default",
				name:      "bar",
			},
			wantErr: true,
		},
		{
			name: "fails when entity config does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
			wantErr: true,
		},
		{
			name: "succeeds when entity config exists",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "default", "bar")
			},
		},
		{
			name: "succeeds when entity config is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "default", "bar")
				deleteEntityConfig(t, s, "default", "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewStoreV2(db))
				}
				s := &EntityConfigStore{
					db: db,
				}
				if err := s.Delete(tt.args.ctx, tt.args.namespace, tt.args.name); (err != nil) != tt.wantErr {
					t.Errorf("EntityConfigStore.Delete() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		})
	}
}

func TestEntityConfigStore_Exists(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace string
		name      string
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
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
		},
		{
			name: "returns false when namespace is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				deleteNamespace(t, s, "default")
			},
		},
		{
			name: "returns false when entity config does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
		},
		{
			name: "returns false when entity config is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "default", "bar")
				deleteEntityConfig(t, s, "default", "bar")
			},
		},
		{
			name: "returns true when entity config exists",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "default", "bar")
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
				s := &EntityConfigStore{
					db: db,
				}
				got, err := s.Exists(tt.args.ctx, tt.args.namespace, tt.args.name)
				if (err != nil) != tt.wantErr {
					t.Errorf("EntityConfigStore.Exists() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("EntityConfigStore.Exists() = %v, want %v", got, tt.want)
				}
			})
		})
	}
}

func TestEntityConfigStore_Get(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace string
		name      string
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.Interface)
		want       *corev3.EntityConfig
		wantErr    bool
	}{
		{
			name: "fails when namespace does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "foo",
			},
			wantErr: true,
		},
		{
			name: "fails when namespace is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "foo",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				deleteNamespace(t, s, "default")
			},
			wantErr: true,
		},
		{
			name: "fails when entity config does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "foo",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
			wantErr: true,
		},
		{
			name: "fails when entity config is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "foo",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "default", "foo")
				deleteEntityConfig(t, s, "default", "foo")
			},
			wantErr: true,
		},
		{
			name: "succeeds when entity config exists",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "foo",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "default", "foo")
			},
			want: func() *corev3.EntityConfig {
				config := corev3.FixtureEntityConfig("foo")
				config.Metadata.Namespace = "default"
				// TODO: uncomment after ID, CreatedAt, UpdatedAt & DeletedAt
				// are added to corev3.EntityConfig.
				//
				// config.ID = 1
				// config.CreatedAt = time.Now()
				// config.UpdatedAt = time.Now()
				return config
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewStoreV2(db))
				}
				s := &EntityConfigStore{
					db: db,
				}
				got, err := s.Get(tt.args.ctx, tt.args.namespace, tt.args.name)
				if (err != nil) != tt.wantErr {
					t.Errorf("EntityConfigStore.Get() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				// TODO: uncomment after ID, CreatedAt, UpdatedAt & DeletedAt
				// are added to corev3.EntityConfig.
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
					t.Errorf("EntityConfigStore.Get() = %v, want %v", got, tt.want)
				}
			})
		})
	}
}

func TestEntityConfigStore_HardDelete(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace string
		name      string
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
				namespace: "default",
				name:      "bar",
			},
			wantErr: true,
		},
		{
			name: "fails when namespace is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				deleteNamespace(t, s, "default")
			},
			wantErr: true,
		},
		{
			name: "fails when entity config does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
			wantErr: true,
		},
		{
			name: "succeeds when entity config is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "default", "bar")
				deleteEntityConfig(t, s, "default", "bar")
			},
		},
		{
			name: "succeeds when entity config exists",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "default", "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewStoreV2(db))
				}
				s := &EntityConfigStore{
					db: db,
				}
				if err := s.HardDelete(tt.args.ctx, tt.args.namespace, tt.args.name); (err != nil) != tt.wantErr {
					t.Errorf("EntityConfigStore.HardDelete() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		})
	}
}

func TestEntityConfigStore_HardDeleted(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace string
		name      string
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
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			want: true,
		},
		{
			name: "returns true when namespace is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				deleteNamespace(t, s, "default")
			},
			want: true,
		},
		{
			name: "returns true when entity config does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
			want: true,
		},
		{
			name: "returns false when entity config is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "default", "bar")
				deleteEntityConfig(t, s, "default", "bar")
			},
			want: false,
		},
		{
			name: "returns false when entity config exists",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "default", "bar")
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
				s := &EntityConfigStore{
					db: db,
				}
				got, err := s.HardDeleted(tt.args.ctx, tt.args.namespace, tt.args.name)
				if (err != nil) != tt.wantErr {
					t.Errorf("EntityConfigStore.HardDeleted() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("EntityConfigStore.HardDeleted() = %v, want %v", got, tt.want)
				}
			})
		})
	}
}

func TestEntityConfigStore_List(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace string
		pred      *store.SelectionPredicate
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.Interface)
		want       []*corev3.EntityConfig
		wantErr    bool
	}{
		{
			name: "fails when namespace does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
			},
			want: nil,
		},
		{
			name: "fails when namespace is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				deleteNamespace(t, s, "default")
			},
			want: nil,
		},
		{
			name: "succeeds when no entity configs exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
			want: nil,
		},
		{
			name: "succeeds when entity configs exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					createEntityConfig(t, s, "default", entityName)
				}
			},
			want: func() []*corev3.EntityConfig {
				configs := []*corev3.EntityConfig{}
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					config := corev3.FixtureEntityConfig(entityName)
					configs = append(configs, config)
				}
				return configs
			}(),
		},
		{
			name: "succeeds when limit set and entity configs exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				pred:      &store.SelectionPredicate{Limit: 5},
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					createEntityConfig(t, s, "default", entityName)
				}
			},
			want: []*corev3.EntityConfig{
				corev3.FixtureEntityConfig("foo-0"),
				corev3.FixtureEntityConfig("foo-1"),
				corev3.FixtureEntityConfig("foo-2"),
				corev3.FixtureEntityConfig("foo-3"),
				corev3.FixtureEntityConfig("foo-4"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewStoreV2(db))
				}
				s := &EntityConfigStore{
					db: db,
				}
				got, err := s.List(tt.args.ctx, tt.args.namespace, tt.args.pred)
				if (err != nil) != tt.wantErr {
					t.Errorf("EntityConfigStore.List() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if diff := deep.Equal(got, tt.want); len(diff) > 0 {
					t.Errorf("EntityConfigStore.List() mismatch: %v", diff)
				}
			})
		})
	}
}

func TestEntityConfigStore_Patch(t *testing.T) {
	type args struct {
		ctx        context.Context
		namespace  string
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
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
				patcher: &patch.Merge{
					MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
				},
			},
			wantErr: true,
		},
		{
			name: "fails when namespace is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
				patcher: &patch.Merge{
					MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
				},
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				deleteNamespace(t, s, "default")
			},
			wantErr: true,
		},
		{
			name: "fails when entity config does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
				patcher: &patch.Merge{
					MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
				},
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
			wantErr: true,
		},
		{
			name: "fails when entity config is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
				patcher: &patch.Merge{
					MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
				},
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "default", "bar")
				deleteEntityConfig(t, s, "default", "bar")
			},
			wantErr: true,
		},
		{
			name: "succeeds when entity config exists",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
				patcher: &patch.Merge{
					MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
				},
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "default", "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewStoreV2(db))
				}
				s := &EntityConfigStore{
					db: db,
				}
				if err := s.Patch(tt.args.ctx, tt.args.namespace, tt.args.name, tt.args.patcher, tt.args.conditions); (err != nil) != tt.wantErr {
					t.Errorf("EntityConfigStore.Patch() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		})
	}
}

func TestEntityConfigStore_UpdateIfExists(t *testing.T) {
	type args struct {
		ctx    context.Context
		config *corev3.EntityConfig
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
				ctx:    context.Background(),
				config: corev3.FixtureEntityConfig("bar"),
			},
			wantErr: true,
		},
		{
			name: "fails when namespace is soft deleted",
			args: args{
				ctx:    context.Background(),
				config: corev3.FixtureEntityConfig("bar"),
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				deleteNamespace(t, s, "default")
			},
			wantErr: true,
		},
		{
			name: "fails when entity config does not exist",
			args: args{
				ctx:    context.Background(),
				config: corev3.FixtureEntityConfig("bar"),
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
			},
			wantErr: true,
		},
		{
			name: "succeeds when entity config is soft deleted",
			args: args{
				ctx:    context.Background(),
				config: corev3.FixtureEntityConfig("bar"),
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "default", "bar")
				deleteEntityConfig(t, s, "default", "bar")
			},
		},
		{
			name: "succeeds when entity config exists",
			args: args{
				ctx:    context.Background(),
				config: corev3.FixtureEntityConfig("bar"),
			},
			beforeHook: func(t *testing.T, s storev2.Interface) {
				createNamespace(t, s, "default")
				createEntityConfig(t, s, "default", "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewStoreV2(db))
				}
				s := &EntityConfigStore{
					db: db,
				}
				if err := s.UpdateIfExists(tt.args.ctx, tt.args.config); (err != nil) != tt.wantErr {
					t.Errorf("EntityConfigStore.UpdateIfExists() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		})
	}
}

// {
// 	name: "multiple entity configs can be retrieved",
// 	args: func() args {
// 		cfg := corev3.FixtureEntityConfig("foo")
// 		req := storev2.NewResourceRequestFromResource(cfg)
// 		return args{
// 			req:     req,
// 			wrapper: WrapEntityConfig(cfg),
// 		}
// 	}(),
// 	reqs: func(t *testing.T, s storev2.Interface) []storev2.ResourceRequest {
// 		createNamespace(t, s, "default")
// 		reqs := make([]storev2.ResourceRequest, 0)
// 		for i := 0; i < 10; i++ {
// 			entityName := fmt.Sprintf("foo-%d", i)
// 			cfg := corev3.FixtureEntityConfig(entityName)
// 			req := storev2.NewResourceRequestFromResource(cfg)
// 			reqs = append(reqs, req)
// 			createEntityConfig(t, s, "default", entityName)
// 		}
// 		return reqs
// 	},
// 	test: func(t *testing.T, wrapper storev2.Wrapper) {
// 		var cfg corev3.EntityConfig
// 		if err := wrapper.UnwrapInto(&cfg); err != nil {
// 			t.Error(err)
// 		}

// 		if got, want := len(cfg.Subscriptions), 2; got != want {
// 			t.Errorf("wrong number of subscriptions, got = %v, want %v", got, want)
// 		}
// 		if got, want := len(cfg.KeepaliveHandlers), 1; got != want {
// 			t.Errorf("wrong number of keepalive handlers, got = %v, want %v", got, want)
// 		}
// 	},
// },
