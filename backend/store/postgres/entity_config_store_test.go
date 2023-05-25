package postgres

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/jackc/pgx/v5/pgxpool"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
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
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		wantErr    bool
	}{
		{
			name: "fails when namespace does not exist",
			args: args{
				ctx:          context.Background(),
				entityConfig: corev3.FixtureEntityConfig("bar"),
			},
			wantErr: true,
		},
		{
			name: "fails when namespace is soft deleted",
			args: args{
				ctx:          context.Background(),
				entityConfig: corev3.FixtureEntityConfig("bar"),
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, _ storev2.EntityConfigStore) {
				createNamespace(t, s, "default")
				deleteNamespace(t, s, "default")
			},
			wantErr: true,
		},
		{
			name: "succeeds when entity config does not exist",
			args: args{
				ctx:          context.Background(),
				entityConfig: corev3.FixtureEntityConfig("bar"),
			},
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, _ storev2.EntityConfigStore) {
				createNamespace(t, s, "default")
			},
		},
		{
			name: "succeeds when entity config is soft deleted",
			args: args{
				ctx:          context.Background(),
				entityConfig: corev3.FixtureEntityConfig("bar"),
			},
			beforeHook: func(t *testing.T, nsStore storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, nsStore, "default")
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
			beforeHook: func(t *testing.T, nsStore storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, nsStore, "default")
				createEntityConfig(t, s, "default", "bar")
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
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		afterHook  func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
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
			wantErr: true,
		},
		{
			name: "fails when namespace is soft deleted",
			args: func() args {
				return args{
					ctx:          context.Background(),
					entityConfig: corev3.FixtureEntityConfig("foo"),
				}
			}(),
			beforeHook: func(t *testing.T, nsStore storev2.NamespaceStore, _ storev2.EntityConfigStore) {
				createNamespace(t, nsStore, "default")
				deleteNamespace(t, nsStore, "default")
			},
			wantErr: true,
		},
		{
			name: "creates when entity config does not exist",
			args: func() args {
				return args{
					ctx:          context.Background(),
					entityConfig: corev3.FixtureEntityConfig("foo"),
				}
			}(),
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, _ storev2.EntityConfigStore) {
				createNamespace(t, s, "default")
			},
			afterHook: func(t *testing.T, _ storev2.NamespaceStore, s storev2.EntityConfigStore) {
				ctx := context.Background()
				config, err := s.Get(ctx, "default", "foo")
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
			beforeHook: func(t *testing.T, nsStore storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, nsStore, "default")
				createEntityConfig(t, s, "default", "foo")
			},
			afterHook: func(t *testing.T, _ storev2.NamespaceStore, s storev2.EntityConfigStore) {
				ctx := context.Background()
				config, err := s.Get(ctx, "default", "foo")
				require.NoError(t, err)
				require.Equal(t, "foo", config.Metadata.Name)
				require.Equal(t, "true", config.Metadata.Annotations["updated"])
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				ns := NewNamespaceStore(db)
				ec := NewEntityConfigStore(db)
				if tt.beforeHook != nil {
					tt.beforeHook(t, ns, ec)
				}
				s := &EntityConfigStore{
					db: db,
				}
				if err := s.CreateOrUpdate(tt.args.ctx, tt.args.entityConfig); (err != nil) != tt.wantErr {
					t.Errorf("EntityConfigStore.CreateOrUpdate() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.afterHook != nil {
					tt.afterHook(t, ns, ec)
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
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
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
			beforeHook: func(t *testing.T, s storev2.NamespaceStore, _ storev2.EntityConfigStore) {
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, s, "default", "bar")
				deleteEntityConfig(t, s, "default", "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
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
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				deleteNamespace(t, ns, "default")
			},
		},
		{
			name: "returns false when entity config does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
			},
		},
		{
			name: "returns false when entity config is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, s, "default", "bar")
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
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				deleteNamespace(t, ns, "default")
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, s, "default", "foo")
			},
			want: func() *corev3.EntityConfig {
				config := corev3.FixtureEntityConfig("foo")
				config.Metadata.Namespace = "default"
				return config
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &EntityConfigStore{
					db: db,
				}
				got, err := s.Get(tt.args.ctx, tt.args.namespace, tt.args.name)
				if (err != nil) != tt.wantErr {
					t.Errorf("EntityConfigStore.Get() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if got != nil {
					purgeIndeterminateStoreLabels(got)
				}

				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("EntityConfigStore.Get() = %v, want %v", got, tt.want)
				}
			})
		})
	}
}

func TestEntityConfigStore_GetMultiple(t *testing.T) {
	type args struct {
		ctx       context.Context
		resources namespacedResourceNames
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		want       uniqueEntityConfigs
		wantErr    bool
	}{
		{
			name: "succeeds when namespace does not exist",
			args: args{
				ctx: context.Background(),
				resources: namespacedResourceNames{
					"default": []string{"foo"},
				},
			},
			want: uniqueEntityConfigs{},
		},
		{
			name: "succeeds when namespace is soft deleted",
			args: args{
				ctx: context.Background(),
				resources: namespacedResourceNames{
					"default": []string{"foo", "bar"},
				},
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				deleteNamespace(t, ns, "default")
			},
			want: uniqueEntityConfigs{},
		},
		{
			name: "succeeds when no requested entity configs exist",
			args: args{
				ctx: context.Background(),
				resources: namespacedResourceNames{
					"default": []string{"foo", "bar"},
					"ops":     []string{"elliot", "mr_robot"},
				},
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				createNamespace(t, ns, "ops")
			},
			want: uniqueEntityConfigs{},
		},
		{
			name: "succeeds when all entity configs are soft deleted",
			args: args{
				ctx: context.Background(),
				resources: namespacedResourceNames{
					"default": []string{"foo", "bar"},
					"ops":     []string{"elliot", "mr_robot"},
				},
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				createNamespace(t, ns, "ops")
				namespacedResourceNames := namespacedResourceNames{
					"default": []string{"foo", "bar"},
					"ops":     []string{"elliot", "mr_robot"},
				}
				for namespace, names := range namespacedResourceNames {
					for _, name := range names {
						createEntityConfig(t, s, namespace, name)
						deleteEntityConfig(t, s, namespace, name)
					}
				}
			},
			want: uniqueEntityConfigs{},
		},
		{
			name: "succeeds when some entity configs are soft deleted",
			args: args{
				ctx: context.Background(),
				resources: namespacedResourceNames{
					"default": []string{"foo", "bar"},
					"ops":     []string{"elliot", "mr_robot"},
				},
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				createNamespace(t, ns, "ops")
				resources := namespacedResourceNames{
					"default": []string{"foo"},
					"ops":     []string{"elliot"},
				}
				for namespace, names := range resources {
					for _, name := range names {
						createEntityConfig(t, s, namespace, name)
					}
				}
				softDeletedResources := namespacedResourceNames{
					"default": []string{"bar"},
					"ops":     []string{"mr_robot"},
				}
				for namespace, names := range softDeletedResources {
					for _, name := range names {
						createEntityConfig(t, s, namespace, name)
						deleteEntityConfig(t, s, namespace, name)
					}
				}
			},
			want: func() uniqueEntityConfigs {
				configs := uniqueEntityConfigs{}
				resources := namespacedResourceNames{
					"default": []string{"foo"},
					"ops":     []string{"elliot"},
				}
				for namespace, names := range resources {
					for _, name := range names {
						resource := uniqueResource{
							Name:      name,
							Namespace: namespace,
						}
						config := corev3.FixtureEntityConfig(name)
						config.Metadata.Namespace = namespace
						configs[resource] = config
					}
				}
				return configs
			}(),
		},
		{
			name: "succeeds when all entity configs exists",
			args: args{
				ctx: context.Background(),
				resources: namespacedResourceNames{
					"default": []string{"foo", "bar"},
					"ops":     []string{"elliot", "mr_robot"},
				},
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				createNamespace(t, ns, "ops")
				resources := namespacedResourceNames{
					"default": []string{"foo", "bar"},
					"ops":     []string{"elliot", "mr_robot"},
				}
				for namespace, names := range resources {
					for _, name := range names {
						createEntityConfig(t, s, namespace, name)
					}
				}
			},
			want: func() uniqueEntityConfigs {
				configs := uniqueEntityConfigs{}
				resources := namespacedResourceNames{
					"default": []string{"foo", "bar"},
					"ops":     []string{"elliot", "mr_robot"},
				}
				for namespace, names := range resources {
					for _, name := range names {
						resource := uniqueResource{
							Name:      name,
							Namespace: namespace,
						}
						config := corev3.FixtureEntityConfig(name)
						config.Metadata.Namespace = namespace
						configs[resource] = config
					}
				}
				return configs
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &EntityConfigStore{
					db: db,
				}
				got, err := s.GetMultiple(tt.args.ctx, tt.args.resources)
				if (err != nil) != tt.wantErr {
					t.Errorf("EntityConfigStore.GetMultiple() error = %v, wantErr %v", err, tt.wantErr)
				}

				for _, g := range got {
					purgeIndeterminateStoreLabels(g)
				}
				if diff := deep.Equal(got, tt.want); diff != nil {
					t.Errorf("EntityConfigStore.GetMultiple() got differs from want: %v", diff)
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
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				deleteNamespace(t, ns, "default")
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, s, "default", "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
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
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				deleteNamespace(t, ns, "default")
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, s, "default", "bar")
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
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				deleteNamespace(t, ns, "default")
			},
			want: nil,
		},
		{
			name: "succeeds when no entity configs exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
			},
			want: nil,
		},
		{
			name: "succeeds when entity configs exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
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
			name: "excludes deleted entity configs",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					createEntityConfig(t, s, "default", entityName)
					if i%2 == 0 {
						if err := s.Delete(context.TODO(), "default", entityName); err != nil {
							t.Fatal(err)
						}
					}
				}

			},
			want: func() []*corev3.EntityConfig {
				configs := []*corev3.EntityConfig{}
				for i := 0; i < 10; i++ {
					if i%2 == 0 {
						continue
					}
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
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
		{
			name: "queries across all namespaces",
			args: args{
				ctx:       context.Background(),
				namespace: "",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "ns0")
				createNamespace(t, ns, "ns1")
				createNamespace(t, ns, "ns2")
				for i := 0; i < 9; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					createEntityConfig(t, s, fmt.Sprintf("ns%d", i%3), entityName)
				}
			},
			want: func() []*corev3.EntityConfig {
				want := make([]*corev3.EntityConfig, 9)
				for i := range want {
					entityName := fmt.Sprintf("foo-%d", i)
					want[i] = corev3.FixtureEntityConfig(entityName)
					want[i].Metadata.Namespace = fmt.Sprintf("ns%d", i%3)
				}
				sort.Slice(want, func(i, j int) bool {
					if want[i].Metadata.Namespace == want[j].Metadata.Namespace {
						return want[i].Metadata.Name < want[j].Metadata.Name
					}
					return want[i].Metadata.Namespace < want[j].Metadata.Namespace
				})
				return want
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &EntityConfigStore{
					db: db,
				}
				got, err := s.List(tt.args.ctx, tt.args.namespace, tt.args.pred)
				if (err != nil) != tt.wantErr {
					t.Errorf("EntityConfigStore.List() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				for _, g := range got {
					purgeIndeterminateStoreLabels(g)
				}
				if diff := deep.Equal(got, tt.want); len(diff) > 0 {
					t.Errorf("EntityConfigStore.List() got differs from want: %v", diff)
				}
			})
		})
	}
}

func TestEntityConfigStore_UpdatedSince(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		s := &EntityConfigStore{
			db: db,
		}
		ns := &NamespaceStore{
			db: db,
		}
		createNamespace(t, ns, "default")
		createEntityConfig(t, s, "default", "older")
		entity, err := s.Get(ctx, "default", "older")
		if err != nil {
			t.Error(err)
			return
		}
		updatedSince := entity.Metadata.Labels[store.SensuUpdatedAtKey]
		if updatedSince == "" {
			t.Error("no updated_since attribute")
			return
		}
		pred := &store.SelectionPredicate{
			UpdatedSince: updatedSince,
		}
		time.Sleep(2 * time.Second) // ensure that "newer" is at least one second older than "older"
		createEntityConfig(t, s, "default", "newer")
		entities, err := s.List(ctx, "default", pred)
		if err != nil {
			t.Error(err)
			return
		}
		if len(entities) != 1 {
			t.Errorf("wrong number of entities: want 1, got %d", len(entities))
			return
		}
		if got, want := entities[0].Metadata.Name, "newer"; got != want {
			t.Errorf("bad entity name: got %q, want %q", got, want)
		}
	})
}

func TestEntityConfigStoreDeletedAt(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		s := &EntityConfigStore{
			db: db,
		}
		ns := &NamespaceStore{
			db: db,
		}
		createNamespace(t, ns, "default")
		createEntityConfig(t, s, "default", "todelete")
		if err := s.Delete(ctx, "default", "todelete"); err != nil {
			t.Error(err)
			return
		}
		pred := &store.SelectionPredicate{
			IncludeDeletes: true,
		}
		entities, err := s.List(ctx, "default", pred)
		if err != nil {
			t.Error(err)
			return
		}
		if len(entities) != 1 {
			t.Errorf("wrong number of entities: want 1, got %d", len(entities))
			return
		}
		if got, want := entities[0].Metadata.Name, "todelete"; got != want {
			t.Errorf("bad entity name: got %q, want %q", got, want)
		}
	})

}

func TestEntityConfigStore_Count(t *testing.T) {
	type args struct {
		ctx         context.Context
		namespace   string
		entityClass string
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
		want       int
		wantErr    bool
	}{
		{
			name: "fails when namespace does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
			},
		},
		{
			name: "fails when namespace is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				deleteNamespace(t, ns, "default")
			},
		},
		{
			name: "succeeds when no entity configs exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
			},
			want: 0,
		},
		{
			name: "succeeds when entity configs exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					createEntityConfig(t, s, "default", entityName)
				}
			},
			want: 10,
		},
		{
			name: "succeeds when entity agent class is specified",
			args: args{
				ctx:         context.Background(),
				namespace:   "default",
				entityClass: corev2.EntityAgentClass,
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					createEntityConfig(t, s, "default", entityName)
				}
			},
			want: 10,
		},
		{
			name: "succeeds when namespace is unspecified",
			args: args{
				ctx:       context.Background(),
				namespace: "",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "ns0")
				createNamespace(t, ns, "ns1")
				createNamespace(t, ns, "ns2")
				for i := 0; i < 9; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					createEntityConfig(t, s, fmt.Sprintf("ns%d", i%3), entityName)
				}
			},
			want: 9,
		},
		{
			name: "succeeds when entity proxy class is specified",
			args: args{
				ctx:         context.Background(),
				namespace:   "default",
				entityClass: corev2.EntityProxyClass,
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					createEntityConfig(t, s, "default", entityName)
				}
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &EntityConfigStore{
					db: db,
				}
				got, err := s.Count(tt.args.ctx, tt.args.namespace, tt.args.entityClass)
				if (err != nil) != tt.wantErr {
					t.Errorf("EntityConfigStore.Count() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if diff := deep.Equal(got, tt.want); len(diff) > 0 {
					t.Errorf("EntityConfigStore.Count() got differs from want: %v", diff)
				}
			})
		})
	}
}

func TestEntityConfigStore_Patch(t *testing.T) {
	type args struct {
		ctx         context.Context
		namespace   string
		name        string
		patcher     patch.Patcher
		ifMatch     storev2.IfMatch
		ifNoneMatch storev2.IfNoneMatch
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				deleteNamespace(t, ns, "default")
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, s, "default", "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
				}
				s := &EntityConfigStore{
					db: db,
				}
				ctx = storev2.ContextWithIfMatch(tt.args.ctx, tt.args.ifMatch)
				ctx = storev2.ContextWithIfNoneMatch(ctx, tt.args.ifNoneMatch)
				if err := s.Patch(ctx, tt.args.namespace, tt.args.name, tt.args.patcher); (err != nil) != tt.wantErr {
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
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore)
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				deleteNamespace(t, ns, "default")
			},
			wantErr: true,
		},
		{
			name: "fails when entity config does not exist",
			args: args{
				ctx:    context.Background(),
				config: corev3.FixtureEntityConfig("bar"),
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
			},
			wantErr: true,
		},
		{
			name: "succeeds when entity config is soft deleted",
			args: args{
				ctx:    context.Background(),
				config: corev3.FixtureEntityConfig("bar"),
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, s storev2.EntityConfigStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, s, "default", "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db))
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
