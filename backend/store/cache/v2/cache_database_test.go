package v2_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/echlebek/migration"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	cachev2 "github.com/sensu/sensu-go/backend/store/cache/v2"
	"github.com/sensu/sensu-go/backend/store/postgres"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

type poolWithDSNFunc func(ctx context.Context, db *pgxpool.Pool, dsn string)

func dropAll(tb testing.TB, dbName, pgURL string) {
	db, err := pgxpool.New(context.Background(), pgURL)
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
	initialDB, err := pgxpool.New(ctx, pgURL)
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
	}, postgres.Migrations)
}

func TestCachingWithPostgresConfigStore(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		s := postgres.NewStore(postgres.StoreConfig{
			DB:     db,
			MaxTPS: 100,
		})
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		cache, err := cachev2.New[*corev2.Asset](ctx, s, false)
		if err != nil {
			t.Error(err)
			return
		}
		watcher := cache.Watch(ctx)
		assetStore := storev2.Of[*corev2.Asset](s)
		for i := 0; i < 10; i++ {
			asset := corev2.FixtureAsset(fmt.Sprintf("%d", i))
			if err := assetStore.CreateIfNotExists(ctx, asset); err != nil {
				t.Error(err)
				return
			}
		}
		<-watcher
		assets := cache.Get("default")
		if got, want := len(assets), 10; got != want {
			t.Errorf("bad assets: got %d, want %d", got, want)
		}
		for i := 0; i < 5; i++ {
			asset := corev2.FixtureAsset(fmt.Sprintf("%d", i+10))
			if err := assetStore.CreateIfNotExists(ctx, asset); err != nil {
				t.Error(err)
				return
			}
		}
		<-watcher
		assets = cache.Get("default")
		if got, want := len(assets), 15; got != want {
			t.Errorf("bad assets: got %d, want %d", got, want)
		}
		for i := 0; i < 5; i++ {
			if err := assetStore.Delete(ctx, storev2.ID{Namespace: "default", Name: fmt.Sprintf("%d", i)}); err != nil {
				t.Error(err)
				return
			}
		}
		<-watcher
		assets = cache.Get("default")
		if got, want := len(assets), 10; got != want {
			t.Errorf("bad assets: got %d, want %d", got, want)
		}
	})
}

func TestCachingWithPostgresEntityConfigStore(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		s := postgres.NewStore(postgres.StoreConfig{
			DB:     db,
			MaxTPS: 100,
		})
		nsStore := storev2.Of[*corev3.Namespace](s)
		ns := &corev3.Namespace{
			Metadata: &corev2.ObjectMeta{
				Name:        "default",
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
		}
		if err := nsStore.CreateIfNotExists(ctx, ns); err != nil {
			t.Error(err)
			return
		}
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		cache, err := cachev2.New[*corev3.EntityConfig](ctx, s, false)
		if err != nil {
			t.Error(err)
			return
		}
		watcher := cache.Watch(ctx)
		entityStore := storev2.Of[*corev3.EntityConfig](s)
		for i := 0; i < 10; i++ {
			entity := corev3.FixtureEntityConfig(fmt.Sprintf("%d", i))
			if err := entityStore.CreateIfNotExists(ctx, entity); err != nil {
				t.Error(err)
				return
			}
		}
		<-watcher
		entities := cache.Get("default")
		if got, want := len(entities), 10; got != want {
			t.Errorf("bad entities: got %d, want %d", got, want)
		}
		for i := 0; i < 5; i++ {
			entity := corev3.FixtureEntityConfig(fmt.Sprintf("%d", i+10))
			if err := entityStore.CreateIfNotExists(ctx, entity); err != nil {
				t.Error(err)
				return
			}
		}
		<-watcher
		entities = cache.Get("default")
		if got, want := len(entities), 15; got != want {
			t.Errorf("bad entities: got %d, want %d", got, want)
		}
		for i := 0; i < 5; i++ {
			if err := entityStore.Delete(ctx, storev2.ID{Namespace: "default", Name: fmt.Sprintf("%d", i)}); err != nil {
				t.Error(err)
				return
			}
		}
		<-watcher
		entities = cache.Get("default")
		if got, want := len(entities), 10; got != want {
			t.Errorf("bad entities: got %d, want %d", got, want)
		}
	})
}
