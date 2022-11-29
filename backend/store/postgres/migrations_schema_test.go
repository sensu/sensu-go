package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const testRefreshUpdatedAtTable = `
CREATE TABLE IF NOT EXISTS test_updated_at (
    name               text NOT NULL,
    updated_at         timestamptz
);
`

const testRefreshUpdatedAtTrigger = `
CREATE TRIGGER test_refresh_updated_at BEFORE UPDATE
    ON test_updated_at FOR EACH ROW
    EXECUTE PROCEDURE refresh_updated_at_column();
`

func TestRefreshUpdatedAtProcedure(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		// create a table & trigger to test the procedure
		if _, err := db.Exec(ctx, testRefreshUpdatedAtTable); err != nil {
			t.Fatalf("error creating table: %v", err)
		}
		if _, err := db.Exec(ctx, testRefreshUpdatedAtTrigger); err != nil {
			t.Fatalf("error creating trigger: %v", err)
		}

		var updatedAt *time.Time

		// insert two rows with no value for updated_at
		names := []string{"foo", "bar"}
		insertQuery := "INSERT INTO test_updated_at (name) VALUES($1)"
		for _, name := range names {
			if _, err := db.Exec(ctx, insertQuery, name); err != nil {
				t.Fatalf("error inserting row: %v", err)
			}
		}

		// ensure the updated_at value is null for both rows
		for _, name := range names {
			query := "SELECT updated_at FROM test_updated_at WHERE name = $1"
			row := db.QueryRow(ctx, query, name)
			if err := row.Scan(&updatedAt); err != nil {
				t.Fatal(err)
			}
			if got, want := updatedAt, (*time.Time)(nil); got != want {
				t.Errorf("updated_at = %v, want %v", got, want)
			}
		}

		// update row
		now := time.Now()
		if _, err := db.Exec(ctx, "UPDATE test_updated_at SET name = 'baz' WHERE name = 'foo'"); err != nil {
			t.Fatalf("error updating row: %v", err)
		}

		// ensure the updated_at value for foo is set
		row := db.QueryRow(ctx, "SELECT updated_at FROM test_updated_at WHERE name = 'baz'")
		if err := row.Scan(&updatedAt); err != nil {
			t.Fatal(err)
		}
		if got, after := updatedAt, now.Add(-1*time.Second); !got.After(after) {
			t.Errorf("updated_at = %v, want >= %v", got, after)
		}
		if got, before := updatedAt, now.Add(1*time.Second); !got.Before(before) {
			t.Errorf("updated_at = %v, want <= %v", got, before)
		}

		// ensure the updated_at value for bar is null
		row = db.QueryRow(ctx, "SELECT updated_at FROM test_updated_at WHERE name = 'bar'")
		if err := row.Scan(&updatedAt); err != nil {
			t.Fatal(err)
		}
		if got, want := updatedAt, (*time.Time)(nil); got != want {
			t.Errorf("updated_at = %v, want %v", got, want)
		}
	})
}
