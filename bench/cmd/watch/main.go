package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/kylelemons/godebug/pretty"
	"github.com/lib/pq"
	"github.com/sensu/sensu-go/bench/watcher"
)

func main() {
	dsn := flag.String("dsn", "", "Postgres connection string")
	flag.Parse()

	testCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdown
		cancel()
	}()

	if dsn == nil || *dsn == "" {
		log.Fatal("Must configure DSN. --dsn 'host=localhost'")
	}

	db, _ := sql.Open("postgres", *dsn)
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to open connection to database: %v", err)
	}

	watcherUnderTest := watcher.Watcher{
		Interval:  time.Millisecond * 1000,
		TxnWindow: time.Millisecond * 5000,
		Head: func(ctx context.Context) (time.Time, error) {
			row := db.QueryRowContext(ctx, "SELECT COALESCE (MAX(updated_at) + '1 microsecond'::interval, NOW()) FROM counters")
			t := time.Time{}
			err := row.Scan(&t)
			fmt.Printf("head, %v\n", t)
			return t, err
		},
		Updates: func(ctx context.Context, t time.Time) ([]watcher.Wrapper, error) {
			var results []watcher.Wrapper
			rows, err := db.QueryContext(ctx, "SELECT id, c, created_at, updated_at, deleted_at from counters WHERE updated_at >= $1", t)
			if err != nil {
				fmt.Printf("updates failed with error %v\n", err)
				return results, err
			}
			defer rows.Close()
			for rows.Next() {
				var id, c int64
				var created, updated time.Time
				var deletedP *time.Time
				var deleted pq.NullTime
				if err := rows.Scan(&id, &c, &created, &updated, &deleted); err != nil {
					fmt.Printf("updates scan failed with error %v\n", err)
					return results, err
				}
				if deleted.Valid {
					deletedP = &deleted.Time
				}
				results = append(results, watcher.Wrapper{Id: fmt.Sprint(id), CreatedAt: created, UpdatedAt: updated, DeletedAt: deletedP, Resource: resource{id, c}})
			}
			err = rows.Close()
			return results, err
		},
	}

	state := getState(testCtx, db)

	events, err := watcherUnderTest.Watch(testCtx)
	if err != nil {
		log.Fatalf("unexpected error watching: %v", err)
	}

	var actualEvents []watcher.Event
	for event := range events {
		r := event.Resource.(resource)
		switch event.Action {
		case watcher.Delete:
			delete(state, r.id)
		case watcher.Update, watcher.Create:
			state[r.id] = r.c
		default:
			log.Fatalf("unexpected action %v", event.Action)
		}
		actualEvents = append(actualEvents, event)
	}
	if len(actualEvents) == 0 {
		log.Fatalf("No changes processed")
	}
	log.Printf("Processed a total of %d events\n", len(actualEvents))
	stateAfter := getState(context.Background(), db)

	if !reflect.DeepEqual(state, stateAfter) {
		log.Println("Expected state does not match actual")
		log.Fatal(pretty.Compare(stateAfter, state))
	}
	log.Println(":thumbsup:")
}

func getState(ctx context.Context, db *sql.DB) map[int64]int64 {
	state := make(map[int64]int64)
	rows, err := db.QueryContext(ctx, "SELECT id, c from counters WHERE deleted_at IS NULL;")
	if err != nil {
		log.Fatalf("error querying counters: %v", err)
	}
	for rows.Next() {
		var id, c int64
		rows.Scan(&id, &c)
		state[id] = c
	}
	if err := rows.Close(); err != nil {
		log.Fatalf("error scanning counters rows: %v", err)
	}
	return state
}

type resource struct {
	id int64
	c  int64
}
