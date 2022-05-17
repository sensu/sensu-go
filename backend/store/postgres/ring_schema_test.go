package postgres

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func TestInsertEntityQuery(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		entities := []string{
			"mulder",
			"scully",
			"skinner",
			"spender",
			"alien bounty hunter",
			"smoking man",
			"teenage vampire",
		}
		for _, entity := range entities {
			row := db.QueryRow(ctx, insertEntityQuery, "default", entity, time.Hour.String())
			var inserted bool
			if err := row.Scan(&inserted); err != nil {
				t.Fatal(err)
			}
			if got, want := inserted, true; got != want {
				t.Errorf("bad inserted: got %v, want %v", got, want)
			}
		}
		// try to insert them all again, this produces conflicts, but not errors
		for _, entity := range entities {
			row := db.QueryRow(ctx, insertEntityQuery, "default", entity, time.Hour.String())
			var inserted bool
			// expect to get ErrNoRows
			if err := row.Scan(&inserted); err != nil {
				t.Fatal(err)
			}
			if got, want := inserted, false; got != want {
				t.Errorf("bad inserted: got %v, want %v", got, want)
			}
		}
		// use a different namespace; this will not produce a conflict
		for _, entity := range entities {
			row := db.QueryRow(ctx, insertEntityQuery, "not-default", entity, time.Hour.String())
			var inserted bool
			if err := row.Scan(&inserted); err != nil {
				t.Fatal(err)
			}
			if got, want := inserted, true; got != want {
				t.Errorf("bad inserted: got %v, want %v", got, want)
			}
		}
	})
}

func TestInsertRingQuery(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		rings := []string{
			"diamond",
			"iron",
			"rosie",
		}
		for _, ring := range rings {
			if _, err := db.Exec(ctx, insertRingQuery, ring); err != nil {
				t.Fatal(err)
			}
		}
		// inserting the rings again should not result in error
		for _, ring := range rings {
			if _, err := db.Exec(ctx, insertRingQuery, ring); err != nil {
				t.Fatal(err)
			}
		}
	})
}

func TestInsertRingEntitiesQuery(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		entities := []string{
			"mulder",
			"scully",
			"skinner",
			"spender",
			"alien bounty hunter",
			"smoking man",
			"teenage vampire",
		}
		for _, entity := range entities {
			row := db.QueryRow(ctx, insertEntityQuery, "default", entity, time.Hour.String())
			var inserted bool
			if err := row.Scan(&inserted); err != nil {
				t.Fatal(err)
			}
		}
		rings := []string{
			"diamond",
			"iron",
			"rosie",
		}
		for _, ring := range rings {
			if _, err := db.Exec(ctx, insertRingQuery, ring); err != nil {
				t.Fatal(err)
			}
		}
		for _, entity := range entities {
			for _, ring := range rings {
				row := db.QueryRow(ctx, insertRingEntityQuery, "default", entity, ring)
				var inserted bool
				if err := row.Scan(&inserted); err != nil {
					t.Fatal(err)
				}
				if got, want := inserted, true; got != want {
					t.Fatalf("bad inserted: got %v, want %v", got, want)
				}
			}
		}
		// We should get errors if the ring doesn't exist
		for _, entity := range entities {
			row := db.QueryRow(ctx, insertRingEntityQuery, "default", entity, "does not exist")
			var inserted bool
			if err := row.Scan(&inserted); err != pgx.ErrNoRows && err.Error() != pgx.ErrNoRows.Error() {
				t.Fatalf("expected pgx.ErrNoRows, got %q (%T)", err, err)
			}
		}
		// We should get errors if the member doesn't exist
		for _, ring := range rings {
			row := db.QueryRow(ctx, insertRingEntityQuery, "default", "does not exist", ring)
			var inserted bool
			if err := row.Scan(&inserted); err != pgx.ErrNoRows && err.Error() != pgx.ErrNoRows.Error() {
				t.Fatalf("expected pgx.ErrNoRows, got %q (%T)", err, err)
			}
		}
	})
}

func TestInsertRingSubscriberQuery(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		entities := []string{
			"mulder",
			"scully",
			"skinner",
			"spender",
			"alien bounty hunter",
			"smoking man",
			"teenage vampire",
		}
		rings := []string{
			"diamond",
			"iron",
			"rosie",
		}
		for _, entity := range entities {
			row := db.QueryRow(ctx, insertEntityQuery, "default", entity, time.Hour.String())
			var inserted bool
			if err := row.Scan(&inserted); err != nil {
				t.Fatal(err)
			}
			for _, ring := range rings {
				if _, err := db.Exec(ctx, insertRingQuery, ring); err != nil {
					t.Fatal(err)
				}
				row := db.QueryRow(ctx, insertRingEntityQuery, "default", entity, ring)
				var inserted bool
				if err := row.Scan(&inserted); err != nil {
					t.Fatal(err)
				}
				if got, want := inserted, true; got != want {
					t.Fatalf("bad inserted: got %v, want %v", got, want)
				}
			}
		}
		row := db.QueryRow(ctx, insertRingSubscriberQuery, "diamond", "my subscriber")
		var inserted bool
		if err := row.Scan(&inserted); err != nil {
			t.Fatal(err)
		}
		// second insert should do nothing
		row = db.QueryRow(ctx, insertRingSubscriberQuery, "diamond", "my subscriber")
		if err := row.Scan(&inserted); err != pgx.ErrNoRows && err.Error() != pgx.ErrNoRows.Error() {
			t.Fatalf("wanted pgx.ErrNoRows, got %#v", err)
		}
	})
}

func TestUpdateRingSubscriberQuery(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		entities := []string{
			"mulder",
			"scully",
			"skinner",
			"spender",
			"alien bounty hunter",
			"smoking man",
			"teenage vampire",
		}
		rings := []string{
			"diamond",
			"iron",
			"rosie",
		}
		for _, entity := range entities {
			row := db.QueryRow(ctx, insertEntityQuery, "default", entity, time.Hour.String())
			var inserted bool
			if err := row.Scan(&inserted); err != nil {
				t.Fatal(err)
			}
			for _, ring := range rings {
				if _, err := db.Exec(ctx, insertRingQuery, ring); err != nil {
					t.Fatal(err)
				}
				row := db.QueryRow(ctx, insertRingEntityQuery, "default", entity, ring)
				var inserted bool
				if err := row.Scan(&inserted); err != nil {
					t.Fatal(err)
				}
				if got, want := inserted, true; got != want {
					t.Fatalf("bad inserted: got %v, want %v", got, want)
				}
			}
		}
		// insert some entities that shouldn't affect anything, due to them being expired
		for _, entity := range []string{"foo", "bar", "baz"} {
			_, err := db.Exec(ctx, insertEntityQuery, "default", entity, (-time.Hour).String())
			if err != nil {
				t.Fatal(err)
			}
			for _, ring := range rings {
				if _, err := db.Exec(ctx, insertRingQuery, ring); err != nil {
					t.Fatal(err)
				}
				_, err := db.Exec(ctx, insertRingEntityQuery, "default", entity, ring)
				if err != nil {
					t.Fatal(err)
				}
			}
		}
		row := db.QueryRow(ctx, insertRingSubscriberQuery, "diamond", "my subscriber")
		var inserted bool
		if err := row.Scan(&inserted); err != nil {
			t.Fatal(err)
		}
		row = db.QueryRow(ctx, insertRingSubscriberQuery, "diamond", "other subscriber")
		if err := row.Scan(&inserted); err != nil {
			t.Fatal(err)
		}
		row = db.QueryRow(ctx, updateRingSubscribersQuery, "diamond", "my subscriber", 0, "0s")
		var entity string
		if err := row.Scan(&entity); err != nil {
			t.Fatal(err)
		}
		// Expect the lexicographically second entity
		if got, want := entity, "mulder"; got != want {
			t.Errorf("bad entity: got %q, want %q", got, want)
		}
		row = db.QueryRow(ctx, `SELECT entities.name FROM entities, ring_subscribers WHERE entities.id = ring_subscribers.pointer AND ring_subscribers.name = 'my subscriber'`)
		var pointer string
		if err := row.Scan(&pointer); err != nil {
			t.Fatal(err)
		}
		if got, want := pointer, "mulder"; got != want {
			t.Errorf("bad pointer: got %q, want %q", got, want)
		}
		pointer = ""
		row = db.QueryRow(ctx, `SELECT entities.name FROM entities, ring_subscribers WHERE ring_subscribers.name = 'other subscriber' AND entities.id = ring_subscribers.pointer`)
		if err := row.Scan(&pointer); err != nil {
			t.Fatal(err)
		}
		if got, want := pointer, "alien bounty hunter"; got != want {
			t.Errorf("bad pointer: got %q, want %q", got, want)
		}
		pointer = ""
		// Iterate a few entities at a time
		row = db.QueryRow(ctx, updateRingSubscribersQuery, "diamond", "my subscriber", 2, "0s")
		if err := row.Scan(&pointer); err != nil {
			t.Fatal(err)
		}
		if got, want := pointer, "smoking man"; got != want {
			t.Errorf("bad pointer: got %q, want %q", got, want)
		}
		row = db.QueryRow(ctx, `SELECT entities.name FROM entities, ring_subscribers WHERE entities.id = ring_subscribers.pointer AND ring_subscribers.name = 'my subscriber'`)
		if err := row.Scan(&pointer); err != nil {
			t.Fatal(err)
		}
		if got, want := pointer, "smoking man"; got != want {
			t.Errorf("bad pointer: got %q, want %q", got, want)
		}
		// Pointer should wrap around at the end
		row = db.QueryRow(ctx, updateRingSubscribersQuery, "diamond", "my subscriber", 1, "0s")
		if err := row.Scan(&pointer); err != nil {
			t.Fatal(err)
		}
		if got, want := pointer, "teenage vampire"; got != want {
			t.Errorf("bad pointer: got %q, want %q", got, want)
		}
		row = db.QueryRow(ctx, `SELECT entities.name FROM entities, ring_subscribers WHERE entities.id = ring_subscribers.pointer AND ring_subscribers.name = 'my subscriber'`)
		if err := row.Scan(&pointer); err != nil {
			t.Fatal(err)
		}
		if got, want := pointer, "teenage vampire"; got != want {
			t.Errorf("bad pointer: got %q, want %q", got, want)
		}
	})
}

func TestUpdateRingSubscriberQueryWithSubsequentEntityInsert(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		entities := []string{
			"mulder",
			"scully",
			"skinner",
			"spender",
			"alien bounty hunter",
			"smoking man",
			"teenage vampire",
		}
		rings := []string{
			"diamond",
			"iron",
			"rosie",
		}
		for _, ring := range rings {
			if _, err := db.Exec(ctx, insertRingQuery, ring); err != nil {
				t.Fatal(err)
			}
		}
		row := db.QueryRow(ctx, insertRingSubscriberQuery, "diamond", "my subscriber")
		var inserted bool
		if err := row.Scan(&inserted); err != nil {
			t.Fatal(err)
		}
		// pointer should **not** be initialized
		row = db.QueryRow(ctx, `SELECT entities.name FROM entities, ring_subscribers WHERE entities.id = ring_subscribers.pointer`)
		var pointer string
		if err := row.Scan(&pointer); err == nil {
			t.Fatal("expected non-nil error")
		}
		for _, entity := range entities {
			row := db.QueryRow(ctx, insertEntityQuery, "default", entity, time.Hour.String())
			var inserted bool
			if err := row.Scan(&inserted); err != nil {
				t.Fatal(err)
			}
			for _, ring := range rings {
				row := db.QueryRow(ctx, insertRingEntityQuery, "default", entity, ring)
				var inserted bool
				if err := row.Scan(&inserted); err != nil {
					t.Fatal(err)
				}
				if got, want := inserted, true; got != want {
					t.Fatalf("bad inserted: got %v, want %v", got, want)
				}
			}
		}
		var entity string
		row = db.QueryRow(ctx, updateRingSubscribersQuery, "diamond", "my subscriber", 0, "1s")
		if err := row.Scan(&entity); err != nil {
			t.Fatal(err)
		}
		if got, want := entity, "alien bounty hunter"; got != want {
			t.Errorf("bad entity: got %q, want %q", got, want)
		}
		row = db.QueryRow(ctx, "SELECT pointer FROM ring_subscribers")
		var ptr int64
		if err := row.Scan(&ptr); err != nil {
			t.Fatal(err)
		}
		if ptr == 0 {
			t.Fatal("ptr is null")
		}
	})
}

func TestGetRingEntitiesQuery(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		entities := []string{
			"mulder",
			"scully",
			"skinner",
			"spender",
			"alien bounty hunter",
			"smoking man",
			"teenage vampire",
		}
		rings := []string{
			"diamond",
			"iron",
			"rosie",
		}
		for _, entity := range entities {
			row := db.QueryRow(ctx, insertEntityQuery, "default", entity, time.Hour.String())
			var inserted bool
			if err := row.Scan(&inserted); err != nil {
				t.Fatal(err)
			}
			for _, ring := range rings {
				if _, err := db.Exec(ctx, insertRingQuery, ring); err != nil {
					t.Fatal(err)
				}
				row := db.QueryRow(ctx, insertRingEntityQuery, "default", entity, ring)
				var inserted bool
				if err := row.Scan(&inserted); err != nil {
					t.Fatal(err)
				}
				if got, want := inserted, true; got != want {
					t.Fatalf("bad inserted: got %v, want %v", got, want)
				}
			}
		}
		row := db.QueryRow(ctx, insertRingSubscriberQuery, "diamond", "my subscriber")
		var inserted bool
		if err := row.Scan(&inserted); err != nil {
			t.Fatal(err)
		}
		var entity string
		row = db.QueryRow(ctx, "SELECT entities.name FROM entities, ring_subscribers WHERE entities.id = ring_subscribers.pointer")
		if err := row.Scan(&entity); err != nil {
			t.Fatal(err)
		}
		if got, want := entity, "alien bounty hunter"; got != want {
			t.Errorf("bad entity: got %q, want %q", got, want)
		}
		rows, err := db.Query(ctx, getRingEntitiesQuery, "diamond", "my subscriber", 2)
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"alien bounty hunter", "mulder"}
		got := []string{}
		for rows.Next() {
			var entity string
			if err := rows.Scan(&entity); err != nil {
				t.Fatal(err)
			}
			got = append(got, entity)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("bad entities: got %v, want %v", got, want)
		}
		row = db.QueryRow(ctx, updateRingSubscribersQuery, "diamond", "my subscriber", 0, "0s")
		if err := row.Scan(&entity); err != nil {
			t.Fatal(err)
		}
		if got, want := entity, "mulder"; got != want {
			t.Errorf("bad entity: got %q, want %q", got, want)
		}
		row = db.QueryRow(ctx, "SELECT entities.name FROM entities, ring_subscribers WHERE entities.id = ring_subscribers.pointer")
		if err := row.Scan(&entity); err != nil {
			t.Fatal(err)
		}
		if got, want := entity, "mulder"; got != want {
			t.Errorf("bad entity: got %q, want %q", got, want)
		}
		rows, err = db.Query(ctx, getRingEntitiesQuery, "diamond", "my subscriber", 2)
		if err != nil {
			t.Fatal(err)
		}
		got = nil
		want = []string{"mulder", "scully"}
		for rows.Next() {
			var entity string
			if err := rows.Scan(&entity); err != nil {
				t.Fatal(err)
			}
			got = append(got, entity)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("bad entities: got %v, want %v", got, want)
		}
		if _, err := db.Exec(ctx, updateRingSubscribersQuery, "diamond", "my subscriber", 1, "0s"); err != nil {
			t.Fatal(err)
		}
		// demonstrate wrap-around
		rows, err = db.Query(ctx, getRingEntitiesQuery, "diamond", "my subscriber", 7)
		if err != nil {
			t.Fatal(err)
		}
		got = nil
		want = []string{"skinner", "smoking man", "spender", "teenage vampire", "alien bounty hunter", "mulder", "scully"}
		for rows.Next() {
			var entity string
			if err := rows.Scan(&entity); err != nil {
				t.Fatal(err)
			}
			got = append(got, entity)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("bad entities: got %v, want %v", got, want)
		}
	})
}

func TestGetRingLengthQuery(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		entities := []string{
			"mulder",
			"scully",
			"skinner",
			"spender",
			"alien bounty hunter",
			"smoking man",
			"teenage vampire",
		}
		rings := []string{
			"diamond",
			"iron",
			"rosie",
		}
		for _, entity := range entities {
			row := db.QueryRow(ctx, insertEntityQuery, "default", entity, time.Hour.String())
			var inserted bool
			if err := row.Scan(&inserted); err != nil {
				t.Fatal(err)
			}
			for _, ring := range rings {
				if _, err := db.Exec(ctx, insertRingQuery, ring); err != nil {
					t.Fatal(err)
				}
				row := db.QueryRow(ctx, insertRingEntityQuery, "default", entity, ring)
				var inserted bool
				if err := row.Scan(&inserted); err != nil {
					t.Fatal(err)
				}
				if got, want := inserted, true; got != want {
					t.Fatalf("bad inserted: got %v, want %v", got, want)
				}
			}
		}
		row := db.QueryRow(ctx, insertRingSubscriberQuery, "diamond", "my subscriber")
		var inserted bool
		if err := row.Scan(&inserted); err != nil {
			t.Fatal(err)
		}
		row = db.QueryRow(ctx, getRingLengthQuery, "diamond")
		var length int
		if err := row.Scan(&length); err != nil {
			t.Fatal(err)
		}
		if got, want := length, len(entities); got != want {
			t.Errorf("bad ring length: got %d, want %d", got, want)
		}
	})
}

func TestDeleteRingEntityQuery(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		entities := []string{
			"mulder",
			"scully",
			"skinner",
			"spender",
			"alien bounty hunter",
			"smoking man",
			"teenage vampire",
		}
		rings := []string{
			"diamond",
			"iron",
			"rosie",
		}
		for _, entity := range entities {
			row := db.QueryRow(ctx, insertEntityQuery, "default", entity, time.Hour.String())
			var inserted bool
			if err := row.Scan(&inserted); err != nil {
				t.Fatal(err)
			}
			for _, ring := range rings {
				if _, err := db.Exec(ctx, insertRingQuery, ring); err != nil {
					t.Fatal(err)
				}
				row := db.QueryRow(ctx, insertRingEntityQuery, "default", entity, ring)
				var inserted bool
				if err := row.Scan(&inserted); err != nil {
					t.Fatal(err)
				}
				if got, want := inserted, true; got != want {
					t.Fatalf("bad inserted: got %v, want %v", got, want)
				}
			}
		}
		row := db.QueryRow(ctx, insertRingSubscriberQuery, "diamond", "my subscriber")
		var inserted bool
		if err := row.Scan(&inserted); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(ctx, deleteRingEntityQuery, "default", "diamond", "teenage vampire"); err != nil {
			t.Fatal(err)
		}
		row = db.QueryRow(ctx, getRingLengthQuery, "diamond")
		var length int
		if err := row.Scan(&length); err != nil {
			t.Fatal(err)
		}
		if got, want := length, len(entities)-1; got != want {
			t.Errorf("bad ring length: got %d, want %d", got, want)
		}
		row = db.QueryRow(ctx, getRingLengthQuery, "iron")
		if err := row.Scan(&length); err != nil {
			t.Fatal(err)
		}
		if got, want := length, len(entities); got != want {
			t.Errorf("bad ring length: got %d, want %d", got, want)
		}
	})
}
