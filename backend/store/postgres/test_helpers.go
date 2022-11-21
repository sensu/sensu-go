package postgres

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/echlebek/migration"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

type poolWithDSNFunc func(ctx context.Context, db *pgxpool.Pool, dsn string)

func dropAll(tb testing.TB, dbName, pgURL string) {
	db, err := pgxpool.Connect(context.Background(), pgURL)
	if err != nil {
		tb.Fatalf("error opening database: %v", err)
	}
	tb.Cleanup(func() {
		db.Close()
	})

	// revoke new connections to the database
	revokeQ := fmt.Sprintf("REVOKE CONNECT ON DATABASE %s FROM public;", dbName)
	if _, err := db.Exec(context.Background(), revokeQ); err != nil {
		tb.Fatalf("error cleaning up database \"%s\", revoke new connections: %v", dbName, err)
	}

	// kill connections to the database
	rawKillQ := `
SELECT pid, pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = '%s' AND pid <> pg_backend_pid();
`
	killQ := fmt.Sprintf(rawKillQ, dbName)
	// this can have permissions errors for reasons unclear to me
	_, _ = db.Exec(context.Background(), killQ)

	// drop the database
	dropQ := fmt.Sprintf("DROP DATABASE %s;", dbName)
	if _, err := db.Exec(context.Background(), dropQ); err != nil {
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
	if err != nil {
		tb.Error(err)
		return
	}
	tb.Cleanup(initialDB.Close)

	id, err := uuid.NewRandom()
	if err != nil {
		tb.Error(err)
	}
	dbName := "sensu" + strings.ReplaceAll(id.String(), "-", "")
	if _, err := initialDB.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s;", dbName)); err != nil {
		tb.Error(err)
		return
	}
	tb.Cleanup(func() {
		dropAll(tb, dbName, pgURL)
	})
	initialDB.Close()

	// connect to postgres again to run migrations
	dsn := fmt.Sprintf("%s dbname=%s", pgURL, dbName)
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		tb.Error(err)
		return
	}

	db, err := migration.Open(cfg, migrations)
	if err != nil {
		tb.Error(err)
		return
	}
	tb.Cleanup(func() { go db.Close() })

	fn(ctx, db, dsn)
}

// creates a database & runs all migrations
func withPostgres(tb testing.TB, fn poolWithDSNFunc) {
	tb.Helper()

	withMigratedPostgres(tb, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		fn(ctx, db, dsn)
	}, migrations)
}

func WithPostgres(tb testing.TB, fn poolWithDSNFunc) {
	withPostgres(tb, fn)
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
func testWithPostgresStore(tb testing.TB, fn func(storev2.Interface)) {
	tb.Helper()

	withPostgres(tb, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		pgStore := &Store{
			db: db,
		}
		ns := &corev3.Namespace{
			Metadata: &corev2.ObjectMeta{
				Name: "default",
			},
		}
		if err := pgStore.GetNamespaceStore().CreateIfNotExists(ctx, ns); err != nil {
			tb.Error(err)
			return
		}
		fn(pgStore)
	})
}

func testWithPostgresEventStore(tb testing.TB, fn func(store.EventStore)) {
	tb.Helper()

	withPostgres(tb, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		pgStore := &Store{
			db: db,
		}
		ns := &corev3.Namespace{
			Metadata: &corev2.ObjectMeta{
				Name: "default",
			},
		}
		if err := pgStore.GetNamespaceStore().CreateIfNotExists(ctx, ns); err != nil {
			tb.Error(err)
			return
		}
		eventStore, err := NewEventStore(db, nil, Config{
			DSN: dsn,
		}, 100)

		if err != nil {
			tb.Error(err)
			return
		}

		fn(eventStore)
	})
}

func createNamespace(tb testing.TB, s storev2.NamespaceStore, name string) {
	tb.Helper()
	ctx := context.Background()
	namespace := corev3.FixtureNamespace(name)
	if err := s.CreateIfNotExists(ctx, namespace); err != nil {
		tb.Error(err)
	}
}

func deleteNamespace(tb testing.TB, s storev2.NamespaceStore, name string) {
	tb.Helper()
	ctx := context.Background()
	if err := s.Delete(ctx, name); err != nil {
		tb.Error(err)
	}
}

func createEntityConfig(tb testing.TB, s storev2.EntityConfigStore, namespace, name string) {
	tb.Helper()
	ctx := context.Background()
	cfg := corev3.FixtureEntityConfig(name)
	cfg.Metadata.Namespace = namespace
	if err := s.CreateIfNotExists(ctx, cfg); err != nil {
		tb.Error(err)
	}
}

func deleteEntityConfig(tb testing.TB, s storev2.EntityConfigStore, namespace, name string) {
	tb.Helper()
	ctx := context.Background()
	if err := s.Delete(ctx, namespace, name); err != nil {
		tb.Error(err)
	}
}

func createEntityState(tb testing.TB, s storev2.EntityStateStore, namespace, name string) {
	tb.Helper()
	ctx := context.Background()
	cfg := corev3.FixtureEntityState(name)
	cfg.Metadata.Namespace = namespace
	if err := s.CreateIfNotExists(ctx, cfg); err != nil {
		tb.Error(err)
	}
}

func deleteEntityState(tb testing.TB, s storev2.EntityStateStore, namespace, name string) {
	tb.Helper()
	ctx := context.Background()
	if err := s.Delete(ctx, namespace, name); err != nil {
		tb.Error(err)
	}
}
