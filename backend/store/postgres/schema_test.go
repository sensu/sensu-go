package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lib/pq"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

func withPostgres(t testing.TB, fn func(context.Context, *pgxpool.Pool, string)) {
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
	defer dropAll(ctx, dbName, pgURL)
	db.Close()
	dsn := fmt.Sprintf("dbname=%s ", dbName) + pgURL
	db, err = pgxpool.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := upgrade(ctx, db); err != nil {
		t.Fatal(err)
	}
	fn(ctx, db, dsn)
}

// only applies the very first schema migration which creates the schema
func withInitialPostgres(t testing.TB, fn func(context.Context, *pgxpool.Pool)) {
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
	defer dropAll(ctx, dbName, pgURL)
	db.Close()
	db, err = pgxpool.Connect(ctx, fmt.Sprintf("dbname=%s ", dbName)+pgURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := upgradeMigration(ctx, db, 0); err != nil {
		t.Fatal(err)
	}
	fn(ctx, db)
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

func upgrade(ctx context.Context, db *pgxpool.Pool) (err error) {
	tx, berr := db.Begin(ctx)
	if berr != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit(ctx)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	for i, migration := range migrations {
		if err := migration(tx); err != nil {
			return fmt.Errorf("migration %d: %s", i, err)
		}
	}
	return nil
}

func dropAll(ctx context.Context, dbName, pgURL string) {
	db, err := sql.Open("postgres", pgURL)
	if err != nil {
		panic(err)
	}
	_, _ = db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE %s;", dbName))
}

func TestUpdateEvents(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		// Create a single event
		row := db.QueryRow(ctx, UpdateEventQuery, "default", "entity", "check", 1, 12345, `{"foo":"bar"}`, []byte("{}"))
		var (
			tsArray, statusArray                      pq.Int64Array
			historyIndex                              int64
			lastOK, occurrences, occurrencesWatermark int64
			previousSerialized                        []byte
		)
		if err := row.Scan(&tsArray, &statusArray, &historyIndex, &lastOK, &occurrences, &occurrencesWatermark, &previousSerialized); err != nil {
			t.Fatal(err)
		}
		if got, want := tsArray, (pq.Int64Array{12345}); !reflect.DeepEqual(got, want) {
			t.Errorf("bad history_ts: got %v, want %v", got, want)
		}
		if got, want := statusArray, (pq.Int64Array{1}); !reflect.DeepEqual(got, want) {
			t.Errorf("bad history_status: got %v, want %v", got, want)
		}
		if got, want := len(previousSerialized), 0; got != want {
			t.Error("expected zero bytes to be returned for serialized")
		}
		if got, want := lastOK, int64(0); got != want {
			t.Errorf("bad last_ok: got %d, want %d", got, want)
		}
		if got, want := occurrences, int64(1); got != want {
			t.Errorf("bad occurrences: got %d, want %d", got, want)
		}
		if got, want := occurrencesWatermark, int64(1); got != want {
			t.Errorf("bad occurrences_wm: got %d, want %d", got, want)
		}

		// Test that the event inserted as expected
		var (
			id, index                int64
			namespace, check, entity string
			serialized               []byte
		)
		row = db.QueryRow(ctx, "SELECT id, sensu_namespace, sensu_check, sensu_entity, history_index, serialized, last_ok, occurrences, occurrences_wm FROM events")
		if err := row.Scan(&id, &namespace, &check, &entity, &index, &serialized, &lastOK, &occurrences, &occurrencesWatermark); err != nil {
			t.Fatal(err)
		}
		if got, want := index, int64(2); got != want {
			t.Errorf("bad index: got %d, want %d", got, want)
		}
		if got, want := namespace, "default"; got != want {
			t.Errorf("bad namespace: got %s, want %s", got, want)
		}
		if got, want := check, "check"; got != want {
			t.Errorf("bad check: got %s, want %s", got, want)
		}
		if got, want := entity, "entity"; got != want {
			t.Errorf("bad entity: got %s, want %s", got, want)
		}
		if got, want := serialized, []byte(`{"foo":"bar"}`); !bytes.Equal(got, want) {
			t.Errorf("bad serialized: got %s, want %s", string(got), string(want))
		}
		if got, want := lastOK, int64(0); got != want {
			t.Errorf("bad last_ok: got %d, want %d", got, want)
		}
		if got, want := occurrences, int64(1); got != want {
			t.Errorf("bad occurrences: got %d, want %d", got, want)
		}
		if got, want := occurrencesWatermark, int64(1); got != want {
			t.Errorf("bad occurrences_wm: got %d, want %d", got, want)
		}

		// Test that the event updates after the first insertion. The row is
		// explicitly selected by the id we generated in the first insert.

		row = db.QueryRow(ctx, UpdateEventQuery, "default", "entity", "check", 0, 54321, `{}`, []byte("{}"))

		if err := row.Scan(&tsArray, &statusArray, &historyIndex, &lastOK, &occurrences, &occurrencesWatermark, &previousSerialized); err != nil {
			t.Fatal(err)
		}
		if got, want := tsArray, (pq.Int64Array{12345, 54321}); !reflect.DeepEqual(got, want) {
			t.Errorf("bad history_ts: got %v, want %v", got, want)
		}
		if got, want := statusArray, (pq.Int64Array{1, 0}); !reflect.DeepEqual(got, want) {
			t.Errorf("bad history_status: got %v, want %v", got, want)
		}
		if got, want := previousSerialized, serialized; !bytes.Equal(got, want) {
			t.Errorf("bad previous_serialized: got %s, want %s", string(got), string(want))
		}
		if got, want := lastOK, int64(54321); got != want {
			t.Errorf("bad last_ok: got %d, want %d", got, want)
		}
		if got, want := occurrences, int64(1); got != want {
			t.Errorf("bad occurrences: got %d, want %d", got, want)
		}
		if got, want := occurrencesWatermark, int64(1); got != want {
			t.Errorf("bad occurrences_wm: got %d, want %d", got, want)
		}

		row = db.QueryRow(ctx, "SELECT sensu_namespace, sensu_check, sensu_entity, history_index, serialized, last_ok, occurrences, occurrences_wm FROM events WHERE id = $1", id)
		if err := row.Scan(&namespace, &check, &entity, &index, &serialized, &lastOK, &occurrences, &occurrencesWatermark); err != nil {
			t.Fatal(err)
		}
		if got, want := index, int64(3); got != want {
			t.Errorf("bad index: got %d, want %d", got, want)
		}
		if got, want := namespace, "default"; got != want {
			t.Errorf("bad namespace: got %s, want %s", got, want)
		}
		if got, want := check, "check"; got != want {
			t.Errorf("bad check: got %s, want %s", got, want)
		}
		if got, want := entity, "entity"; got != want {
			t.Errorf("bad entity: got %s, want %s", got, want)
		}
		if got, want := serialized, []byte(`{}`); !bytes.Equal(got, want) {
			t.Errorf("bad serialized: got %s, want %s", string(got), string(want))
		}
		if got, want := lastOK, int64(54321); got != want {
			t.Errorf("bad last_ok: got %d, want %d", got, want)
		}
		if got, want := occurrences, int64(1); got != want {
			t.Errorf("bad occurrences: got %d, want %d", got, want)
		}
		if got, want := occurrencesWatermark, int64(1); got != want {
			t.Errorf("bad occurrences_wm: got %d, want %d", got, want)
		}

		// Test that the index wraps around. This enables a ringbuffer-like
		// behaviour for the history and timestamp arrays. The arrays should
		// never grow to more than 21 elements. This is hardcoded, now, but could
		// change one day.
		for i := 0; i < 25; i++ {
			if _, err := db.Exec(ctx, UpdateEventQuery, "default", "entity", "check", 0, 54321+i, `{}`, []byte("{}")); err != nil {
				t.Fatal(err)
			}
		}
		row = db.QueryRow(ctx, "SELECT sensu_namespace, sensu_check, sensu_entity, history_index, serialized, last_ok, occurrences, occurrences_wm FROM events WHERE id = $1", id)
		if err := row.Scan(&namespace, &check, &entity, &index, &serialized, &lastOK, &occurrences, &occurrencesWatermark); err != nil {
			t.Fatal(err)
		}
		if got, want := index, int64(7); got != want {
			t.Errorf("bad index: got %d, want %d", got, want)
		}
		if got, want := namespace, "default"; got != want {
			t.Errorf("bad namespace: got %s, want %s", got, want)
		}
		if got, want := check, "check"; got != want {
			t.Errorf("bad check: got %s, want %s", got, want)
		}
		if got, want := entity, "entity"; got != want {
			t.Errorf("bad entity: got %s, want %s", got, want)
		}
		if got, want := serialized, []byte(`{}`); !bytes.Equal(got, want) {
			t.Errorf("bad serialized: got %s, want %s", string(got), string(want))
		}
		if got, want := lastOK, int64(54345); got != want {
			t.Errorf("bad last_ok: got %d, want %d", got, want)
		}
		if got, want := occurrences, int64(26); got != want {
			t.Errorf("bad occurrences: got %d, want %d", got, want)
		}
		if got, want := occurrencesWatermark, int64(26); got != want {
			t.Errorf("bad occurrences_wm: got %d, want %d", got, want)
		}

		// Test that LastOK and occurrences are as expected after a set of failures
		if _, err := db.Exec(ctx, UpdateEventQuery, "default", "entity", "check", 1, 0xbeef, `{}`, []byte("{}")); err != nil {
			t.Fatal(err)
		}
		row = db.QueryRow(ctx, "SELECT sensu_namespace, sensu_check, sensu_entity, history_index, serialized, last_ok, occurrences, occurrences_wm FROM events WHERE id = $1", id)
		if err := row.Scan(&namespace, &check, &entity, &index, &serialized, &lastOK, &occurrences, &occurrencesWatermark); err != nil {
			t.Fatal(err)
		}
		if got, want := lastOK, int64(54345); got != want {
			t.Errorf("bad last_ok: got %d, want %d", got, want)
		}
		if got, want := occurrences, int64(1); got != want {
			t.Errorf("bad occurrences: got %d, want %d", got, want)
		}
		if got, want := occurrencesWatermark, int64(1); got != want {
			t.Errorf("bad occurrences_wm: got %d, want %d", got, want)
		}
		if _, err := db.Exec(ctx, UpdateEventQuery, "default", "entity", "check", 1, 0xbeef, `{}`, []byte("{}")); err != nil {
			t.Fatal(err)
		}
		row = db.QueryRow(ctx, "SELECT sensu_namespace, sensu_check, sensu_entity, history_index, serialized, last_ok, occurrences, occurrences_wm FROM events WHERE id = $1", id)
		if err := row.Scan(&namespace, &check, &entity, &index, &serialized, &lastOK, &occurrences, &occurrencesWatermark); err != nil {
			t.Fatal(err)
		}
		if got, want := lastOK, int64(54345); got != want {
			t.Errorf("bad last_ok: got %d, want %d", got, want)
		}
		if got, want := occurrences, int64(2); got != want {
			t.Errorf("bad occurrences: got %d, want %d", got, want)
		}
		if got, want := occurrencesWatermark, int64(2); got != want {
			t.Errorf("bad occurrences_wm: got %d, want %d", got, want)
		}
		if _, err := db.Exec(ctx, UpdateEventQuery, "default", "entity", "check", 2, 0xbeef, `{}`, []byte("{}")); err != nil {
			t.Fatal(err)
		}
		row = db.QueryRow(ctx, "SELECT sensu_namespace, sensu_check, sensu_entity, history_index, serialized, last_ok, occurrences, occurrences_wm FROM events WHERE id = $1", id)
		if err := row.Scan(&namespace, &check, &entity, &index, &serialized, &lastOK, &occurrences, &occurrencesWatermark); err != nil {
			t.Fatal(err)
		}
		if got, want := lastOK, int64(54345); got != want {
			t.Errorf("bad last_ok: got %d, want %d", got, want)
		}
		if got, want := occurrences, int64(1); got != want {
			t.Errorf("bad occurrences: got %d, want %d", got, want)
		}
		if got, want := occurrencesWatermark, int64(2); got != want {
			t.Errorf("bad occurrences_wm: got %d, want %d", got, want)
		}
	})
}

func TestUpdateEventOnly(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		// Create event with UpdateEventQuery
		var (
			id                       int64
			namespace, check, entity string
			serialized               []byte
		)
		if _, err := db.Exec(ctx, UpdateEventQuery, "default", "entity", "check", 3, 12345, `{"foo":"bar"}`, []byte("{}")); err != nil {
			t.Fatal(err)
		}
		row := db.QueryRow(ctx, "SELECT id, sensu_namespace, sensu_check, sensu_entity, serialized FROM events")
		if err := row.Scan(&id, &namespace, &check, &entity, &serialized); err != nil {
			t.Fatal(err)
		}
		if got, want := namespace, "default"; got != want {
			t.Errorf("bad namespace: got %s, want %s", got, want)
		}
		if got, want := check, "check"; got != want {
			t.Errorf("bad check: got %s, want %s", got, want)
		}
		if got, want := entity, "entity"; got != want {
			t.Errorf("bad entity: got %s, want %s", got, want)
		}
		if got, want := serialized, []byte(`{"foo":"bar"}`); !bytes.Equal(got, want) {
			t.Errorf("bad serialized: got %s, want %s", string(got), string(want))
		}

		// Update serialized data with UpdateEventOnlyQuery
		if _, err := db.Exec(ctx, UpdateEventOnlyQuery, []byte(`{"bar":"baz"}`), []byte(`{}`), "default", "entity", "check"); err != nil {
			t.Fatal(err)
		}
		row = db.QueryRow(ctx, "SELECT sensu_namespace, sensu_check, sensu_entity, serialized FROM events WHERE id = $1", id)
		if err := row.Scan(&namespace, &check, &entity, &serialized); err != nil {
			t.Fatal(err)
		}
		if got, want := namespace, "default"; got != want {
			t.Errorf("bad namespace: got %s, want %s", got, want)
		}
		if got, want := check, "check"; got != want {
			t.Errorf("bad check: got %s, want %s", got, want)
		}
		if got, want := entity, "entity"; got != want {
			t.Errorf("bad entity: got %s, want %s", got, want)
		}
		if got, want := serialized, []byte(`{"bar":"baz"}`); !bytes.Equal(got, want) {
			t.Errorf("bad serialized: got %s, want %s", string(got), string(want))
		}
	})
}

func TestGetEventCountsByNamespaces(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		addEvents := func(n int, namespace, checkBaseName string, status int) {
			for i := 0; i < n; i++ {
				checkName := fmt.Sprintf("%s-%d", checkBaseName, i)
				if _, err := db.Exec(context.Background(), UpdateEventQuery, namespace, "entity", checkName, status, 12345, `{"foo":"bar"}`, []byte("{}")); err != nil {
					t.Fatal(err)
				}
			}
		}

		addEvents(5, "default", "check-status-ok", 0)
		addEvents(2, "default", "check-status-warning", 1)

		addEvents(2, "dumpster", "check-status-critical", 2)
		addEvents(2, "dumpster", "check-status-other", 3)

		tests := []struct {
			namespace string
			want      EventGauges
		}{
			{
				namespace: "default",
				want: EventGauges{
					Total:          7,
					StatusOK:       5,
					StatusWarning:  2,
					StatusCritical: 0,
					StatusOther:    0,
				},
			},
			{
				namespace: "dumpster",
				want: EventGauges{
					Total:          4,
					StatusOK:       0,
					StatusWarning:  0,
					StatusCritical: 2,
					StatusOther:    2,
				},
			},
		}

		for _, test := range tests {
			rows, err := db.Query(ctx, GetEventCountsByNamespaceQuery)
			if err != nil {
				t.Fatal(err)
			}

			result, err := scanCounts(rows)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(result[test.namespace], test.want) {
				t.Errorf("got %#v, want %#v", result[test.namespace], test.want)
			}
		}
	})
}

func TestGetKeepaliveCountsByNamespaces(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		addKeepalives := func(n int, namespace, entityBaseName string, status int) {
			for i := 0; i < n; i++ {
				entityName := fmt.Sprintf("%s-%d", entityBaseName, i)
				if _, err := db.Exec(context.Background(), UpdateEventQuery, namespace, entityName, "keepalive", status, 12345, `{"foo":"bar"}`, []byte("{}")); err != nil {
					t.Fatal(err)
				}
			}
		}

		addKeepalives(5, "default", "status-ok", 0)
		addKeepalives(2, "default", "status-warning", 1)

		addKeepalives(2, "dumpster", "status-critical", 2)
		addKeepalives(2, "dumpster", "status-other", 3)

		tests := []struct {
			namespace string
			want      KeepaliveGauges
		}{
			{
				namespace: "default",
				want: KeepaliveGauges{
					Total:          7,
					StatusOK:       5,
					StatusWarning:  2,
					StatusCritical: 0,
					StatusOther:    0,
				},
			},
			{
				namespace: "dumpster",
				want: KeepaliveGauges{
					Total:          4,
					StatusOK:       0,
					StatusWarning:  0,
					StatusCritical: 2,
					StatusOther:    2,
				},
			},
		}

		for _, test := range tests {
			rows, err := db.Query(ctx, GetEventCountsByNamespaceQuery)
			if err != nil {
				t.Fatal(err)
			}

			result, err := scanCounts(rows)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(result[test.namespace], test.want) {
				t.Errorf("got %#v, want %#v", result[test.namespace], test.want)
			}
		}
	})
}

func TestEventStoreSelectorsMigration(t *testing.T) {
	pgURL := os.Getenv("PG_URL")
	if pgURL == "" {
		pgURL = "host=/run/postgresql sslmode=disable"
	}
	withInitialPostgres(t, func(ctx context.Context, db *pgxpool.Pool) {
		if err := upgradeMigration(ctx, db, 3); err != nil {
			t.Fatal(err)
		}
		st, err := NewEventStore(db, nil, Config{
			DSN: pgURL,
		}, 1)
		if err != nil {
			t.Fatal(err)
		}
		ctx = store.NamespaceContext(ctx, "default")
		event := corev2.FixtureEvent("entity", "check")
		for i := 0; i < 1500; i++ {
			event.Check.Name = fmt.Sprintf("%d", i)
			if _, _, err := st.UpdateEvent(ctx, event); err != nil {
				t.Fatal(err)
			}
		}
		if err := upgradeMigration(ctx, db, 4); err != nil {
			t.Fatal(err)
		}
		var result int
		row := db.QueryRow(ctx, `SELECT count(*) FROM events WHERE selectors @> '{"event.check.status": "0"}'`)
		if err := row.Scan(&result); err != nil {
			t.Fatal(err)
		}
		if got, want := result, 1500; got != want {
			t.Errorf("bad result: got %q, want %q", got, want)
		}
	})
}

func benchmarkMigrateSelectors(b *testing.B, n int) {
	pgURL := os.Getenv("PG_URL")
	if pgURL == "" {
		pgURL = "host=/run/postgresql sslmode=disable"
	}
	withInitialPostgres(b, func(ctx context.Context, db *pgxpool.Pool) {
		if err := upgradeMigration(ctx, db, 3); err != nil {
			b.Fatal(err)
		}
		st, err := NewEventStore(db, nil, Config{
			DSN: pgURL,
		}, 1)
		if err != nil {
			b.Fatal(err)
		}
		ctx = store.NamespaceContext(ctx, "default")
		event := corev2.FixtureEvent("entity", "check")
		for i := 0; i < n; i++ {
			event.Check.Name = fmt.Sprintf("%d", i)
			if _, _, err := st.UpdateEvent(ctx, event); err != nil {
				b.Fatal(err)
			}
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tx, err := db.Begin(ctx)
			if err != nil {
				b.Fatal(err)
			}
			if err := migrateUpdateSelectors(tx); err != nil {
				_ = tx.Rollback(ctx)
				b.Fatal(err)
			}
			if err := tx.Commit(ctx); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkMigrateSelectors1000(b *testing.B) {
	benchmarkMigrateSelectors(b, 1000)
}

func BenchmarkMigrateSelectors10000(b *testing.B) {
	benchmarkMigrateSelectors(b, 10000)
}

func BenchmarkMigrateSelectors100000(b *testing.B) {
	benchmarkMigrateSelectors(b, 100000)
}
