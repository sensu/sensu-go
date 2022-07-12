package postgres

import (
	"context"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetOutput(ioutil.Discard)
}

func TestInsertEntityQuery(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)
		entityNames := []string{
			"mulder",
			"scully",
			"skinner",
			"spender",
			"alien bounty hunter",
			"smoking man",
			"teenage vampire",
		}
		for _, entityName := range entityNames {
			entity := corev2.FixtureEntity(entityName)
			namespace := corev2.FixtureNamespace(entity.Namespace)
			if err := namespaceStore.UpdateNamespace(ctx, namespace); err != nil {
				t.Fatal(err)
			}
			if err := entityStore.UpdateEntity(ctx, entity); err != nil {
				t.Fatal(err)
			}
			_, err := db.Exec(ctx, updateEntityStateExpiresAtQuery, "default", entityName, time.Hour.String())
			if err != nil {
				t.Fatal(err)
			}
		}
	})
}

func TestInsertRingQuery(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)
		rings := []string{
			"diamond",
			"iron",
			"rosie",
		}
		for _, ring := range rings {
			entity := corev2.FixtureEntity(ring)
			namespace := corev2.FixtureNamespace(entity.Namespace)
			if err := namespaceStore.UpdateNamespace(ctx, namespace); err != nil {
				t.Fatal(err)
			}
			if err := entityStore.UpdateEntity(ctx, entity); err != nil {
				t.Fatal(err)
			}
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
		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)
		entityNames := []string{
			"mulder",
			"scully",
			"skinner",
			"spender",
			"alien bounty hunter",
			"smoking man",
			"teenage vampire",
		}
		for _, entityName := range entityNames {
			entity := corev2.FixtureEntity(entityName)
			namespace := corev2.FixtureNamespace(entity.Namespace)
			if err := namespaceStore.UpdateNamespace(ctx, namespace); err != nil {
				t.Fatal(err)
			}
			if err := entityStore.UpdateEntity(ctx, entity); err != nil {
				t.Fatal(err)
			}
			_, err := db.Exec(ctx, updateEntityStateExpiresAtQuery, entity.Namespace, entity.Name, time.Hour.String())
			if err != nil {
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
		for _, entityName := range entityNames {
			for _, ring := range rings {
				row := db.QueryRow(ctx, insertRingEntityQuery, "default", entityName, ring)
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
		for _, entityName := range entityNames {
			row := db.QueryRow(ctx, insertRingEntityQuery, "default", entityName, "does not exist")
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
		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)
		entityNames := []string{
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
		for _, entityName := range entityNames {
			entity := corev2.FixtureEntity(entityName)
			namespace := corev2.FixtureNamespace(entity.Namespace)
			if err := namespaceStore.UpdateNamespace(ctx, namespace); err != nil {
				t.Fatal(err)
			}
			if err := entityStore.UpdateEntity(ctx, entity); err != nil {
				t.Fatal(err)
			}
			_, err := db.Exec(ctx, updateEntityStateExpiresAtQuery, entity.Namespace, entity.Name, time.Hour.String())
			if err != nil {
				t.Fatal(err)
			}
			for _, ring := range rings {
				if _, err := db.Exec(ctx, insertRingQuery, ring); err != nil {
					t.Fatal(err)
				}
				row := db.QueryRow(ctx, insertRingEntityQuery, "default", entityName, ring)
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
		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)
		entityNames := []string{
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
		for _, entityName := range entityNames {
			entity := corev2.FixtureEntity(entityName)
			namespace := corev2.FixtureNamespace(entity.Namespace)
			if err := namespaceStore.UpdateNamespace(ctx, namespace); err != nil {
				t.Fatal(err)
			}
			if err := entityStore.UpdateEntity(ctx, entity); err != nil {
				t.Fatal(err)
			}
			_, err := db.Exec(ctx, updateEntityStateExpiresAtQuery, entity.Namespace, entity.Name, time.Hour.String())
			if err != nil {
				t.Fatal(err)
			}
			for _, ring := range rings {
				if _, err := db.Exec(ctx, insertRingQuery, ring); err != nil {
					t.Fatal(err)
				}
				row := db.QueryRow(ctx, insertRingEntityQuery, entity.Namespace, entity.Name, ring)
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
			_, err := db.Exec(ctx, updateEntityStateExpiresAtQuery, "default", entity, (-time.Hour).String())
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
		row = db.QueryRow(ctx, `SELECT entity_states.name FROM entity_states, ring_subscribers WHERE entity_states.id = ring_subscribers.pointer AND ring_subscribers.name = 'my subscriber'`)
		var pointer string
		if err := row.Scan(&pointer); err != nil {
			t.Fatal(err)
		}
		if got, want := pointer, "mulder"; got != want {
			t.Errorf("bad pointer: got %q, want %q", got, want)
		}
		pointer = ""
		row = db.QueryRow(ctx, `SELECT entity_states.name FROM entity_states, ring_subscribers WHERE ring_subscribers.name = 'other subscriber' AND entity_states.id = ring_subscribers.pointer`)
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
		row = db.QueryRow(ctx, `SELECT entity_states.name FROM entity_states, ring_subscribers WHERE entity_states.id = ring_subscribers.pointer AND ring_subscribers.name = 'my subscriber'`)
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
		row = db.QueryRow(ctx, `SELECT entity_states.name FROM entity_states, ring_subscribers WHERE entity_states.id = ring_subscribers.pointer AND ring_subscribers.name = 'my subscriber'`)
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
		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)
		entityNames := []string{
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
		for _, entityName := range entityNames {
			entity := corev2.FixtureEntity(entityName)
			namespace := corev2.FixtureNamespace(entity.Namespace)
			if err := namespaceStore.UpdateNamespace(ctx, namespace); err != nil {
				t.Fatal(err)
			}
			if err := entityStore.UpdateEntity(ctx, entity); err != nil {
				t.Fatal(err)
			}
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
		row = db.QueryRow(ctx, `SELECT entity_states.name FROM entity_states, ring_subscribers WHERE entity_states.id = ring_subscribers.pointer`)
		var pointer string
		if err := row.Scan(&pointer); err == nil {
			t.Fatal("expected non-nil error")
		}
		for _, entityName := range entityNames {
			entity := corev2.FixtureEntity(entityName)
			_, err := db.Exec(ctx, updateEntityStateExpiresAtQuery, entity.Namespace, entity.Name, time.Hour.String())
			if err != nil {
				t.Fatal(err)
			}
			for _, ring := range rings {
				row := db.QueryRow(ctx, insertRingEntityQuery, entity.Namespace, entity.Name, ring)
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
		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)
		entityNames := []string{
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
		for _, entityName := range entityNames {
			entity := corev2.FixtureEntity(entityName)
			namespace := corev2.FixtureNamespace(entity.Namespace)
			if err := namespaceStore.UpdateNamespace(ctx, namespace); err != nil {
				t.Fatal(err)
			}
			if err := entityStore.UpdateEntity(ctx, entity); err != nil {
				t.Fatal(err)
			}
			_, err := db.Exec(ctx, updateEntityStateExpiresAtQuery, entity.Namespace, entity.Name, time.Hour.String())
			if err != nil {
				t.Fatal(err)
			}
			for _, ring := range rings {
				if _, err := db.Exec(ctx, insertRingQuery, ring); err != nil {
					t.Fatal(err)
				}
				row := db.QueryRow(ctx, insertRingEntityQuery, entity.Namespace, entity.Name, ring)
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
		row = db.QueryRow(ctx, "SELECT entity_states.name FROM entity_states, ring_subscribers WHERE entity_states.id = ring_subscribers.pointer")
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
		row = db.QueryRow(ctx, "SELECT entity_states.name FROM entity_states, ring_subscribers WHERE entity_states.id = ring_subscribers.pointer")
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
		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)
		entityNames := []string{
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
		for _, entityName := range entityNames {
			entity := corev2.FixtureEntity(entityName)
			namespace := corev2.FixtureNamespace(entity.Namespace)
			if err := namespaceStore.UpdateNamespace(ctx, namespace); err != nil {
				t.Fatal(err)
			}
			if err := entityStore.UpdateEntity(ctx, entity); err != nil {
				t.Fatal(err)
			}
			_, err := db.Exec(ctx, updateEntityStateExpiresAtQuery, entity.Namespace, entity.Name, time.Hour.String())
			if err != nil {
				t.Fatal(err)
			}
			for _, ring := range rings {
				if _, err := db.Exec(ctx, insertRingQuery, ring); err != nil {
					t.Fatal(err)
				}
				row := db.QueryRow(ctx, insertRingEntityQuery, entity.Namespace, entity.Name, ring)
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
		if got, want := length, len(entityNames); got != want {
			t.Errorf("bad ring length: got %d, want %d", got, want)
		}
	})
}

func TestDeleteRingEntityQuery(t *testing.T) {
	t.Parallel()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		namespaceStore := NewNamespaceStore(db)
		entityStore := NewEntityStore(db)
		entityNames := []string{
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
		for _, entityName := range entityNames {
			entity := corev2.FixtureEntity(entityName)
			namespace := corev2.FixtureNamespace(entity.Namespace)
			if err := namespaceStore.UpdateNamespace(ctx, namespace); err != nil {
				t.Fatal(err)
			}
			if err := entityStore.UpdateEntity(ctx, entity); err != nil {
				t.Fatal(err)
			}
			_, err := db.Exec(ctx, updateEntityStateExpiresAtQuery, entity.Namespace, entity.Name, time.Hour.String())
			if err != nil {
				t.Fatal(err)
			}
			for _, ring := range rings {
				if _, err := db.Exec(ctx, insertRingQuery, ring); err != nil {
					t.Fatal(err)
				}
				row := db.QueryRow(ctx, insertRingEntityQuery, entity.Namespace, entity.Name, ring)
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
		if got, want := length, len(entityNames)-1; got != want {
			t.Errorf("bad ring length: got %d, want %d", got, want)
		}
		row = db.QueryRow(ctx, getRingLengthQuery, "iron")
		if err := row.Scan(&length); err != nil {
			t.Fatal(err)
		}
		if got, want := length, len(entityNames); got != want {
			t.Errorf("bad ring length: got %d, want %d", got, want)
		}
	})
}
