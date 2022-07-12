package postgres

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/jackc/pgx/v4/pgxpool"
	v2 "github.com/sensu/sensu-go/backend/store/v2"
)

func TestCounterWatcherIntegration(t *testing.T) {
	// warn: High Flap potential
	//
	// Set ExpectedTxnJitter to a very high percentile of expected transaction latency.
	// This test case runs ~15k transactions - so p99 would mean on average 150 transactions
	// that could fall out of the watcher's transaction window and cause a failure.
	ExpectedTxnJitter := time.Millisecond * 1000
	// additional transaction commit delay
	SyntheticTxnJitter := time.Millisecond * 50
	TxnWindow := ExpectedTxnJitter + SyntheticTxnJitter

	withPostgres(t, func(ctx context.Context, pool *pgxpool.Pool, dsn string) {
		if _, err := pool.Exec(ctx, counterDDL); err != nil {
			t.Fatalf("could not apply DDL: %v", err)
		}
		counterState, err := queryState(ctx, pool)
		if err != nil {
			t.Fatal(err)
		}

		// Component under test - use watcher to keep counterState in sync with db
		watcherUnderTest := NewStoreV2(pool)
		watcherUnderTest.watchInterval = time.Millisecond * 10
		watcherUnderTest.watchTxnWindow = TxnWindow

		watchCtx, watchCancel := context.WithCancel(ctx)
		defer watchCancel()
		watchDone := watchState(t, watchCtx, watcherUnderTest, counterState)
		<-time.After(time.Millisecond * 10)

		// seed db with counters
		seed(t, watchCtx, pool)
		// generate traffic on counters table
		updateCtx, updateCancel := context.WithCancel(ctx)
		defer updateCancel()
		updateDone := generateCountersTraffic(updateCtx, pool, 32, SyntheticTxnJitter)

		// Allow to run for 10 seconds then cancel and wait for traffic to stop
		<-time.After(time.Second * 10)
		updateCancel()
		<-updateDone

		// wait a full polling interval before stopping watcher
		<-time.After(time.Millisecond * 20)
		watchCancel()
		<-watchDone

		actualState, err := queryState(ctx, pool)
		if err != nil {
			t.Fatal(err)
		}

		if diff := deep.Equal(actualState, counterState); diff != nil {
			t.Fatalf("expected database and local state with watcher to match: %v", diff)
		}
	})
}

func generateCountersTraffic(ctx context.Context, pool *pgxpool.Pool, workers int, jitter time.Duration) <-chan struct{} {
	done := make(chan struct{})

	worker := func(ctx context.Context, pool *pgxpool.Pool) {
		update := "UPDATE counters SET c = c + 1 WHERE id = $1"
		for {
			if ctx.Err() != nil {
				return
			}
			tx, err := pool.Begin(ctx)
			if err != nil {
				return
			}
			err = func(ctx context.Context) (txErr error) {
				defer func() {
					if txErr != nil {
						tx.Rollback(ctx)
						return
					}
					txErr = tx.Commit(ctx)
				}()
				id := rand.Intn(1000)
				duration := jitter * time.Duration(rand.Intn(100)) / 100
				if _, err := tx.Exec(ctx, update, id); err != nil {
					return err
				}
				// wait delay or until ctx cancel
				select {
				case <-time.After(duration):
				case <-ctx.Done():
					return
				}
				return

			}(ctx)
			if err != nil {
				return
			}
		}
	}
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(ctx context.Context) {
			defer wg.Done()
			worker(ctx, pool)
		}(ctx)
	}
	go func() {
		wg.Wait()
		close(done)
	}()
	return done
}

func watchState(t *testing.T, ctx context.Context, w *StoreV2, state map[int64]int64) <-chan struct{} {
	t.Helper()

	done := make(chan struct{})
	watchEvents := w.Watch(ctx, v2.ResourceRequest{
		StoreName: "testing::counter",
	})
	go func() {
		defer func() {
			close(done)
		}()
		for {
			watchEvents, ok := <-watchEvents
			if !ok {
				return
			}
			for _, watchEvent := range watchEvents {
				var r counter
				wrapper := watchEvent.Value
				if wrapper != nil {
					wrapper.UnwrapInto(&r)
				}
				switch watchEvent.Type {
				case v2.WatchDelete:
					delete(state, r.Id)
				case v2.WatchUpdate, v2.WatchCreate:
					state[r.Id] = r.C
				case v2.WatchError:
					if ctx.Err() == nil {
						t.Error("Received watch error")
					}
					return
				default:
					t.Errorf("unexpected action %v", watchEvent.Type)
					return
				}
			}
		}
	}()
	return done
}

func seed(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	insert := "INSERT INTO counters (created_at) SELECT NOW() FROM generate_series(1, 1000);"
	if _, err := pool.Exec(ctx, insert); err != nil {
		t.Errorf("error seeding counters: %v", err)
	}
}

func queryState(ctx context.Context, pool *pgxpool.Pool) (map[int64]int64, error) {
	actualState := make(map[int64]int64)

	rows, err := pool.Query(ctx, "SELECT id, c FROM counters")
	if err != nil {
		return actualState, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, c int64
		if err := rows.Scan(&id, &c); err != nil {
			return actualState, nil
		}
		actualState[id] = c
	}
	return actualState, nil
}
