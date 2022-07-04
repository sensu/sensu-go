package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/echlebek/migration"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/store"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/stretchr/testify/require"
)

type poolWithDSNFunc func(ctx context.Context, db *pgxpool.Pool, dsn string)

func dropAll(tb testing.TB, dbName, pgURL string) {
	db, err := sql.Open("postgres", pgURL)
	if err != nil {
		tb.Fatalf("error opening database: %v", err)
	}
	dropQ := fmt.Sprintf("DROP DATABASE %s;", dbName)
	if _, err := db.ExecContext(context.Background(), dropQ); err != nil {
		tb.Fatalf("error cleaning up database \"%s\": %v", dbName, err)
	}
}

// creates a new database and runs any provided migrations
func withMigratedPostgres(tb testing.TB, fn poolWithDSNFunc, migrations []migration.Migrator) {
	tb.Helper()
	if testing.Short() {
		tb.Skip("skipping postgres test: short mode enabled")
	}
	pgURL := os.Getenv("PG_URL")
	if pgURL == "" {
		tb.Skip("skipping postgres test: PG_URL not set")
	}

	ctx, cancel := context.WithCancel(context.Background())
	tb.Cleanup(cancel)

	// connect to postgres to create the database for tests
	initialDB, err := pgxpool.Connect(ctx, pgURL)
	require.NoError(tb, err)
	tb.Cleanup(initialDB.Close)

	dbName := "sensu" + strings.ReplaceAll(uuid.New().String(), "-", "")
	if _, err := initialDB.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s;", dbName)); err != nil {
		tb.Error(err)
	}
	tb.Cleanup(func() {
		dropAll(tb, dbName, pgURL)
	})
	initialDB.Close()

	// connect to postgres again to run migrations
	dsn := fmt.Sprintf("dbname=%s ", dbName) + pgURL
	cfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(tb, err)

	db, err := migration.Open(cfg, migrations)
	require.NoError(tb, err)
	tb.Cleanup(db.Close)

	fn(ctx, db, dsn)
}

// creates a database & runs all migrations
func withPostgres(tb testing.TB, fn poolWithDSNFunc) {
	tb.Helper()

	withMigratedPostgres(tb, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		fn(ctx, db, dsn)
	}, migrations)
}

// creates a database & only applies the very first schema migration which
// creates the schema
func withInitialPostgres(tb testing.TB, fn func(context.Context, *pgxpool.Pool)) {
	tb.Helper()

	migrations := []migration.Migrator{
		migrations[0],
	}

	withMigratedPostgres(tb, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		fn(ctx, db)
	}, migrations)
}

func upgradeMigration(ctx context.Context, db *pgxpool.Pool, migration int) (err error) {
	tx, berr := db.Begin(ctx)
	if berr != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit(ctx)
		}
		if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				panic(err)
			}
		}
	}()
	if err := migrations[migration](tx); err != nil {
		return err
	}
	return nil
}

// creates a database, runs all migrations & provides a StoreV1
func testWithPostgresStore(tb testing.TB, fn func(store.Store)) {
	tb.Helper()

	withPostgres(tb, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		eventStore, err := NewEventStore(db, nil, Config{
			DSN: dsn,
		}, 100)
		require.NoError(tb, err)

		e, cleanup := etcd.NewTestEtcd(tb)
		tb.Cleanup(cleanup)

		client := e.NewEmbeddedClient()
		etcdStore := etcdstore.NewStore(client)

		if err := etcdStore.CreateNamespace(
			context.Background(),
			corev2.FixtureNamespace("default"),
		); err != nil {
			tb.Fatal(err)
		}

		namespaceStore := NewNamespaceStore(db, client)
		entityStore := NewEntityStore(db, client)

		pgStore := Store{
			EventStore:     eventStore,
			EntityStore:    entityStore,
			NamespaceStore: namespaceStore,
			Store:          etcdStore,
		}
		fn(pgStore)
	})
}

// creates a database, runs all migrations & provides a StoreV2
func testWithPostgresStoreV2(tb testing.TB, fn func(storev2.Interface)) {
	tb.Helper()

	withPostgres(tb, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		fn(NewStoreV2(db, nil))
	})
}

func createNamespace(tb testing.TB, s storev2.Interface, name string) {
	tb.Helper()
	ctx := context.Background()
	namespace := corev3.FixtureNamespace(name)
	req := storev2.NewResourceRequestFromResource(ctx, namespace)
	req.UsePostgres = true
	wrapper := WrapNamespace(namespace)
	if err := s.CreateOrUpdate(req, wrapper); err != nil {
		tb.Error(err)
	}
}

func createEntityConfig(tb testing.TB, s storev2.Interface, name string) {
	tb.Helper()
	ctx := context.Background()
	cfg := corev3.FixtureEntityConfig(name)
	req := storev2.NewResourceRequestFromResource(ctx, cfg)
	req.UsePostgres = true
	wrapper := WrapEntityConfig(cfg)
	if err := s.CreateOrUpdate(req, wrapper); err != nil {
		tb.Error(err)
	}
}

func createEntityState(tb testing.TB, s storev2.Interface, name string) {
	tb.Helper()
	ctx := context.Background()
	state := corev3.FixtureEntityState(name)
	req := storev2.NewResourceRequestFromResource(ctx, state)
	req.UsePostgres = true
	wrapper := WrapEntityState(state)
	if err := s.CreateOrUpdate(req, wrapper); err != nil {
		tb.Error(err)
	}
}
