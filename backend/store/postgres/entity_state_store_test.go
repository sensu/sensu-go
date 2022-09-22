package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-test/deep"
	"github.com/jackc/pgx/v4/pgxpool"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/stretchr/testify/require"
)

func TestEntityStateStore_CreateIfNotExists(t *testing.T) {
	type args struct {
		ctx         context.Context
		entityState *corev3.EntityState
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore, storev2.EntityStateStore)
		wantErr    bool
	}{
		{
			name: "fails when namespace does not exist",
			args: args{
				ctx:         context.Background(),
				entityState: corev3.FixtureEntityState("bar"),
			},
			wantErr: true,
		},
		{
			name: "fails when namespace is soft deleted",
			args: args{
				ctx:         context.Background(),
				entityState: corev3.FixtureEntityState("bar"),
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				deleteNamespace(t, ns, "default")
			},
			wantErr: true,
		},
		{
			name: "fails when entity config does not exist",
			args: args{
				ctx:         context.Background(),
				entityState: corev3.FixtureEntityState("bar"),
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
			},
			wantErr: true,
		},
		{
			name: "fails when entity config is soft deleted",
			args: args{
				ctx:         context.Background(),
				entityState: corev3.FixtureEntityState("bar"),
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				deleteEntityConfig(t, ec, "default", "bar")
			},
			wantErr: true,
		},
		{
			name: "succeeds when entity state does not exist",
			args: args{
				ctx:         context.Background(),
				entityState: corev3.FixtureEntityState("bar"),
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
			},
		},
		{
			name: "succeeds when entity state is soft deleted",
			args: args{
				ctx:         context.Background(),
				entityState: corev3.FixtureEntityState("bar"),
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				createEntityState(t, s, "default", "bar")
				deleteEntityState(t, s, "default", "bar")
			},
		},
		{
			name: "fails when entity state exists",
			args: args{
				ctx:         context.Background(),
				entityState: corev3.FixtureEntityState("bar"),
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				createEntityState(t, s, "default", "bar")
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db), NewEntityStateStore(db))
				}
				s := &EntityStateStore{
					db: db,
				}
				if err := s.CreateIfNotExists(tt.args.ctx, tt.args.entityState); (err != nil) != tt.wantErr {
					t.Errorf("EntityStateStore.CreateIfNotExists() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		})
	}
}

func TestEntityStateStore_CreateOrUpdate(t *testing.T) {
	type args struct {
		ctx         context.Context
		entityState *corev3.EntityState
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore, storev2.EntityStateStore)
		afterHook  func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore, storev2.EntityStateStore)
		wantErr    bool
	}{
		{
			name: "fails when namespace does not exist",
			args: func() args {
				return args{
					ctx:         context.Background(),
					entityState: corev3.FixtureEntityState("foo"),
				}
			}(),
			wantErr: true,
		},
		{
			name: "fails when namespace is soft deleted",
			args: func() args {
				return args{
					ctx:         context.Background(),
					entityState: corev3.FixtureEntityState("foo"),
				}
			}(),
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				deleteNamespace(t, ns, "default")
			},
			wantErr: true,
		},
		{
			name: "fails when entity config does not exist",
			args: func() args {
				return args{
					ctx:         context.Background(),
					entityState: corev3.FixtureEntityState("foo"),
				}
			}(),
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
			},
			wantErr: true,
		},
		{
			name: "fails when entity config is soft deleted",
			args: func() args {
				return args{
					ctx:         context.Background(),
					entityState: corev3.FixtureEntityState("foo"),
				}
			}(),
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "foo")
				deleteEntityConfig(t, ec, "default", "foo")
			},
			wantErr: true,
		},
		{
			name: "creates when entity state does not exist",
			args: func() args {
				return args{
					ctx:         context.Background(),
					entityState: corev3.FixtureEntityState("foo"),
				}
			}(),
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "foo")
			},
			afterHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				ctx := context.Background()
				state, err := s.Get(ctx, "default", "foo")
				require.NoError(t, err)
				require.Equal(t, "foo", state.Metadata.Name)
			},
		},
		{
			name: "updates when entity state exists",
			args: func() args {
				return args{
					ctx: context.Background(),
					entityState: func() *corev3.EntityState {
						state := corev3.FixtureEntityState("foo")
						state.Metadata.Annotations["updated"] = "true"
						return state
					}(),
				}
			}(),
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "foo")
				createEntityState(t, s, "default", "foo")
			},
			afterHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				ctx := context.Background()
				state, err := s.Get(ctx, "default", "foo")
				require.NoError(t, err)
				require.Equal(t, "foo", state.Metadata.Name)
				require.Equal(t, "true", state.Metadata.Annotations["updated"])
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db), NewEntityStateStore(db))
				}
				s := &EntityStateStore{
					db: db,
				}
				if err := s.CreateOrUpdate(tt.args.ctx, tt.args.entityState); (err != nil) != tt.wantErr {
					t.Errorf("EntityStateStore.CreateOrUpdate() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.afterHook != nil {
					tt.afterHook(t, NewNamespaceStore(db), NewEntityConfigStore(db), NewEntityStateStore(db))
				}
			})
		})
	}
}

func TestEntityStateStore_Delete(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace string
		name      string
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore, storev2.EntityStateStore)
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
			},
			wantErr: true,
		},
		{
			name: "fails when entity state does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "foo")
			},
			wantErr: true,
		},
		{
			name: "succeeds when entity state exists",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				createEntityState(t, s, "default", "bar")
			},
		},
		{
			name: "succeeds when entity state is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				createEntityState(t, s, "default", "bar")
				deleteEntityState(t, s, "default", "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db), NewEntityStateStore(db))
				}
				s := &EntityStateStore{
					db: db,
				}
				if err := s.Delete(tt.args.ctx, tt.args.namespace, tt.args.name); (err != nil) != tt.wantErr {
					t.Errorf("EntityStateStore.Delete() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		})
	}
}

func TestEntityStateStore_Exists(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace string
		name      string
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore, storev2.EntityStateStore)
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
			},
		},
		{
			name: "returns false when entity config is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				deleteEntityConfig(t, ec, "default", "bar")
			},
		},
		{
			name: "returns false when entity state does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
			},
		},
		{
			name: "returns false when entity state is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				createEntityState(t, s, "default", "bar")
				deleteEntityState(t, s, "default", "bar")
			},
		},
		{
			name: "returns true when entity state exists",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				createEntityState(t, s, "default", "bar")
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db), NewEntityStateStore(db))
				}
				s := &EntityStateStore{
					db: db,
				}
				got, err := s.Exists(tt.args.ctx, tt.args.namespace, tt.args.name)
				if (err != nil) != tt.wantErr {
					t.Errorf("EntityStateStore.Exists() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("EntityStateStore.Exists() = %v, want %v", got, tt.want)
				}
			})
		})
	}
}

func TestEntityStateStore_Get(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace string
		name      string
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore, storev2.EntityStateStore)
		want       *corev3.EntityState
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "foo")
				deleteEntityConfig(t, ec, "default", "foo")
			},
			wantErr: true,
		},
		{
			name: "fails when entity state does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "foo",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "foo")
			},
			wantErr: true,
		},
		{
			name: "fails when entity state is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "foo",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "foo")
				createEntityState(t, s, "default", "foo")
				deleteEntityState(t, s, "default", "foo")
			},
			wantErr: true,
		},
		{
			name: "succeeds when entity state exists",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "foo",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "foo")
				createEntityState(t, s, "default", "foo")
			},
			want: func() *corev3.EntityState {
				state := corev3.FixtureEntityState("foo")
				state.Metadata.Namespace = "default"
				state.System.Processes = nil
				// TODO: uncomment after ID, CreatedAt, UpdatedAt & DeletedAt
				// are added to corev3.EntityState.
				//
				// state.ID = 1
				// state.CreatedAt = time.Now()
				// state.UpdatedAt = time.Now()
				return state
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db), NewEntityStateStore(db))
				}
				s := &EntityStateStore{
					db: db,
				}
				got, err := s.Get(tt.args.ctx, tt.args.namespace, tt.args.name)
				if (err != nil) != tt.wantErr {
					t.Errorf("EntityStateStore.Get() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				// TODO: uncomment after ID, CreatedAt, UpdatedAt & DeletedAt
				// are added to corev3.EntityState.
				//
				// createdAtDelta := time.Since(tt.want.CreatedAt) / 2
				// wantCreatedAt := time.Now().Add(-createdAtDelta)
				// require.WithinDuration(t, wantCreatedAt, got.CreatedAt, createdAtDelta)
				// got.CreatedAt = tt.want.CreatedAt

				// updatedAtDelta := time.Since(tt.want.UpdatedAt) / 2
				// wantUpdatedAt := time.Now().Add(-updatedAtDelta)
				// require.WithinDuration(t, wantUpdatedAt, got.UpdatedAt, updatedAtDelta)
				// got.UpdatedAt = tt.want.UpdatedAt

				if diff := deep.Equal(got, tt.want); len(diff) != 0 {
					t.Errorf("EntityStateStore.Get() got differs from want = %v", diff)
				}
			})
		})
	}
}

func TestEntityStateStore_GetMultiple(t *testing.T) {
	type args struct {
		ctx       context.Context
		resources namespacedResourceNames
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore, storev2.EntityStateStore)
		want       uniqueEntityStates
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
			want: uniqueEntityStates{},
		},
		{
			name: "succeeds when namespace is soft deleted",
			args: args{
				ctx: context.Background(),
				resources: namespacedResourceNames{
					"default": []string{"foo", "bar"},
				},
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				deleteNamespace(t, ns, "default")
			},
			want: uniqueEntityStates{},
		},
		{
			name: "succeeds when entity config does not exist",
			args: args{
				ctx: context.Background(),
				resources: namespacedResourceNames{
					"default": []string{"foo"},
				},
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
			},
			want: uniqueEntityStates{},
		},
		{
			name: "succeeds when entity config is soft deleted",
			args: args{
				ctx: context.Background(),
				resources: namespacedResourceNames{
					"default": []string{"foo"},
				},
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				deleteEntityConfig(t, ec, "default", "bar")
			},
			want: uniqueEntityStates{},
		},
		{
			name: "succeeds when no requested entity states exist",
			args: args{
				ctx: context.Background(),
				resources: namespacedResourceNames{
					"default": []string{"foo", "bar"},
					"ops":     []string{"elliot", "mr_robot"},
				},
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createNamespace(t, ns, "ops")
				createEntityConfig(t, ec, "default", "foo")
				createEntityConfig(t, ec, "ops", "elliot")
			},
			want: uniqueEntityStates{},
		},
		{
			name: "succeeds when all entity states are soft deleted",
			args: args{
				ctx: context.Background(),
				resources: namespacedResourceNames{
					"default": []string{"foo", "bar"},
					"ops":     []string{"elliot", "mr_robot"},
				},
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createNamespace(t, ns, "ops")
				namespacedResourceNames := namespacedResourceNames{
					"default": []string{"foo", "bar"},
					"ops":     []string{"elliot", "mr_robot"},
				}
				for namespace, names := range namespacedResourceNames {
					for _, name := range names {
						createEntityConfig(t, ec, namespace, name)
						createEntityState(t, s, namespace, name)
						deleteEntityState(t, s, namespace, name)
					}
				}
			},
			want: uniqueEntityStates{},
		},
		{
			name: "succeeds when some entity states are soft deleted",
			args: args{
				ctx: context.Background(),
				resources: namespacedResourceNames{
					"default": []string{"foo", "bar"},
					"ops":     []string{"elliot", "mr_robot"},
				},
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createNamespace(t, ns, "ops")
				resources := namespacedResourceNames{
					"default": []string{"foo"},
					"ops":     []string{"elliot"},
				}
				for namespace, names := range resources {
					for _, name := range names {
						createEntityConfig(t, ec, namespace, name)
						createEntityState(t, s, namespace, name)
					}
				}
				softDeletedResources := namespacedResourceNames{
					"default": []string{"bar"},
					"ops":     []string{"mr_robot"},
				}
				for namespace, names := range softDeletedResources {
					for _, name := range names {
						createEntityConfig(t, ec, namespace, name)
						createEntityState(t, s, namespace, name)
						deleteEntityState(t, s, namespace, name)
					}
				}
			},
			want: func() uniqueEntityStates {
				states := uniqueEntityStates{}
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
						state := corev3.FixtureEntityState(name)
						state.Metadata.Namespace = namespace
						state.System.Processes = nil
						states[resource] = state
					}
				}
				return states
			}(),
		},
		{
			name: "succeeds when all entity states exists",
			args: args{
				ctx: context.Background(),
				resources: namespacedResourceNames{
					"default": []string{"foo", "bar"},
					"ops":     []string{"elliot", "mr_robot"},
				},
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createNamespace(t, ns, "ops")
				resources := namespacedResourceNames{
					"default": []string{"foo", "bar"},
					"ops":     []string{"elliot", "mr_robot"},
				}
				for namespace, names := range resources {
					for _, name := range names {
						createEntityConfig(t, ec, namespace, name)
						createEntityState(t, s, namespace, name)
					}
				}
			},
			want: func() uniqueEntityStates {
				states := uniqueEntityStates{}
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
						state := corev3.FixtureEntityState(name)
						state.Metadata.Namespace = namespace
						state.System.Processes = nil
						states[resource] = state
					}
				}
				return states
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db), NewEntityStateStore(db))
				}
				s := &EntityStateStore{
					db: db,
				}
				got, err := s.GetMultiple(tt.args.ctx, tt.args.resources)
				if (err != nil) != tt.wantErr {
					t.Errorf("EntityStateStore.GetMultiple() error = %v, wantErr %v", err, tt.wantErr)
				}
				if diff := deep.Equal(got, tt.want); diff != nil {
					t.Errorf("EntityStateStore.GetMultiple() got differs from want: %v", diff)
				}
			})
		})
	}
}

func TestEntityStateStore_HardDelete(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace string
		name      string
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore, storev2.EntityStateStore)
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
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
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				deleteEntityConfig(t, ec, "default", "bar")
			},
			wantErr: true,
		},
		{
			name: "fails when entity state does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
			},
			wantErr: true,
		},
		{
			name: "succeeds when entity state is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				createEntityState(t, s, "default", "bar")
				deleteEntityState(t, s, "default", "bar")
			},
		},
		{
			name: "succeeds when entity state exists",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				createEntityState(t, s, "default", "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db), NewEntityStateStore(db))
				}
				s := &EntityStateStore{
					db: db,
				}
				if err := s.HardDelete(tt.args.ctx, tt.args.namespace, tt.args.name); (err != nil) != tt.wantErr {
					t.Errorf("EntityStateStore.HardDelete() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		})
	}
}

func TestEntityStateStore_HardDeleted(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace string
		name      string
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore, storev2.EntityStateStore)
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
			},
			want: true,
		},
		{
			name: "returns true when entity config is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				deleteEntityConfig(t, ec, "default", "bar")
			},
			want: true,
		},
		{
			name: "returns true when entity state does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
			},
			want: true,
		},
		{
			name: "returns false when entity state is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				createEntityState(t, s, "default", "bar")
				deleteEntityState(t, s, "default", "bar")
			},
			want: false,
		},
		{
			name: "returns false when entity state exists",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				createEntityState(t, s, "default", "bar")
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db), NewEntityStateStore(db))
				}
				s := &EntityStateStore{
					db: db,
				}
				got, err := s.HardDeleted(tt.args.ctx, tt.args.namespace, tt.args.name)
				if (err != nil) != tt.wantErr {
					t.Errorf("EntityStateStore.HardDeleted() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("EntityStateStore.HardDeleted() = %v, want %v", got, tt.want)
				}
			})
		})
	}
}

func TestEntityStateStore_List(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace string
		pred      *store.SelectionPredicate
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore, storev2.EntityStateStore)
		want       []*corev3.EntityState
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				deleteNamespace(t, ns, "default")
			},
			want: nil,
		},
		{
			name: "succeeds when no entity states exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					createEntityConfig(t, ec, "default", entityName)
				}
			},
			want: nil,
		},
		{
			name: "succeeds when entity states exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					createEntityConfig(t, ec, "default", entityName)
					createEntityState(t, s, "default", entityName)
				}
			},
			want: func() []*corev3.EntityState {
				states := []*corev3.EntityState{}
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					state := corev3.FixtureEntityState(entityName)
					state.System.Processes = nil
					states = append(states, state)
				}
				return states
			}(),
		},
		{
			name: "succeeds when limit set and entity states exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				pred:      &store.SelectionPredicate{Limit: 5},
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				for i := 0; i < 10; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					createEntityConfig(t, ec, "default", entityName)
					createEntityState(t, s, "default", entityName)
				}
			},
			want: func() []*corev3.EntityState {
				states := []*corev3.EntityState{}
				for i := 0; i < 5; i++ {
					entityName := fmt.Sprintf("foo-%d", i)
					state := corev3.FixtureEntityState(entityName)
					state.System.Processes = nil
					states = append(states, state)
				}
				return states
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db), NewEntityStateStore(db))
				}
				s := &EntityStateStore{
					db: db,
				}
				got, err := s.List(tt.args.ctx, tt.args.namespace, tt.args.pred)
				if (err != nil) != tt.wantErr {
					t.Errorf("EntityStateStore.List() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if diff := deep.Equal(got, tt.want); diff != nil {
					t.Errorf("EntityStateStore.List() got differs from want: %v", diff)
				}
			})
		})
	}
}

func TestEntityStateStore_Patch(t *testing.T) {
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
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore, storev2.EntityStateStore)
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
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
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				deleteEntityConfig(t, ec, "default", "bar")
			},
			wantErr: true,
		},
		{
			name: "fails when entity state does not exist",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
				patcher: &patch.Merge{
					MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
				},
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
			},
			wantErr: true,
		},
		{
			name: "fails when entity state is soft deleted",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
				patcher: &patch.Merge{
					MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
				},
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				createEntityState(t, s, "default", "bar")
				deleteEntityState(t, s, "default", "bar")
			},
			wantErr: true,
		},
		{
			name: "succeeds when entity state exists",
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				name:      "bar",
				patcher: &patch.Merge{
					MergePatch: []byte(`{"metadata":{"labels":{"food":"hummus"}}}`),
				},
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				createEntityState(t, s, "default", "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db), NewEntityStateStore(db))
				}
				s := &EntityStateStore{
					db: db,
				}
				if err := s.Patch(tt.args.ctx, tt.args.namespace, tt.args.name, tt.args.patcher, tt.args.conditions); (err != nil) != tt.wantErr {
					t.Errorf("EntityStateStore.Patch() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		})
	}
}

func TestEntityStateStore_UpdateIfExists(t *testing.T) {
	type args struct {
		ctx   context.Context
		state *corev3.EntityState
	}
	tests := []struct {
		name       string
		args       args
		beforeHook func(*testing.T, storev2.NamespaceStore, storev2.EntityConfigStore, storev2.EntityStateStore)
		wantErr    bool
	}{
		{
			name: "fails when namespace does not exist",
			args: args{
				ctx:   context.Background(),
				state: corev3.FixtureEntityState("bar"),
			},
			wantErr: true,
		},
		{
			name: "fails when namespace is soft deleted",
			args: args{
				ctx:   context.Background(),
				state: corev3.FixtureEntityState("bar"),
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				deleteNamespace(t, ns, "default")
			},
			wantErr: true,
		},
		{
			name: "fails when entity config does not exist",
			args: args{
				ctx:   context.Background(),
				state: corev3.FixtureEntityState("bar"),
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
			},
			wantErr: true,
		},
		{
			name: "fails when entity config is soft deleted",
			args: args{
				ctx:   context.Background(),
				state: corev3.FixtureEntityState("bar"),
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				deleteEntityConfig(t, ec, "default", "bar")
			},
			wantErr: true,
		},
		{
			name: "fails when entity state does not exist",
			args: args{
				ctx:   context.Background(),
				state: corev3.FixtureEntityState("bar"),
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
			},
			wantErr: true,
		},
		{
			name: "succeeds when entity state is soft deleted",
			args: args{
				ctx:   context.Background(),
				state: corev3.FixtureEntityState("bar"),
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				createEntityState(t, s, "default", "bar")
				deleteEntityState(t, s, "default", "bar")
			},
		},
		{
			name: "succeeds when entity state exists",
			args: args{
				ctx:   context.Background(),
				state: corev3.FixtureEntityState("bar"),
			},
			beforeHook: func(t *testing.T, ns storev2.NamespaceStore, ec storev2.EntityConfigStore, s storev2.EntityStateStore) {
				createNamespace(t, ns, "default")
				createEntityConfig(t, ec, "default", "bar")
				createEntityState(t, s, "default", "bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
				if tt.beforeHook != nil {
					tt.beforeHook(t, NewNamespaceStore(db), NewEntityConfigStore(db), NewEntityStateStore(db))
				}
				s := &EntityStateStore{
					db: db,
				}
				if err := s.UpdateIfExists(tt.args.ctx, tt.args.state); (err != nil) != tt.wantErr {
					os.Exit(1)
					t.Errorf("EntityStateStore.UpdateIfExists() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		})
	}
}
