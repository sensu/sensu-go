package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
)

func TestSynchronizedExecutor(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		executor := &SynchronizedExecutor{
			DB:              db,
			CheckinInterval: time.Millisecond * 20,
		}

		err := executor.Execute(ctx, store.Mutex(333), func(ctx context.Context) error { return errors.New("error A") })
		assert.EqualError(t, err, "error A")
		err = executor.Execute(ctx, store.Mutex(333), func(ctx context.Context) error { return errors.New("error B") })
		assert.EqualError(t, err, "error B")

		// test concurrent executions of different mutexes
		step := make(chan struct{}, 1)
		go executor.Execute(ctx, store.Mutex(333), func(ctx context.Context) error {
			<-step
			return nil
		})

		done := make(chan struct{})
		go executor.Execute(ctx, store.Mutex(222), func(ctx context.Context) error {
			close(done)
			return nil
		})

		select {
		case <-done:
			// OK
		case <-time.After(time.Millisecond * 100):
			t.Error("expected concurrent executions of different mutexes")
		}

		// test concurrent executions of the same mutex
		done = make(chan struct{})
		go executor.Execute(ctx, store.Mutex(333), func(ctx context.Context) error {
			close(done)
			return nil
		})

		select {
		case <-done:
			t.Error("expected concurrent execution of busy mutex to block")
		case <-time.After(time.Millisecond * 100):
			// OK
		}

		// unblock first execution
		step <- struct{}{}
		select {
		case <-done:
			// OK
		case <-time.After(time.Millisecond * 100):
			t.Error("expected the second handler to take over after the first completed")
		}

		isLocked := make(chan struct{})
		// test many queued executions
		// heuristic to ensure many executors do not exhaust the connection pool
		go executor.Execute(ctx, store.Mutex(333), func(ctx context.Context) error {
			close(isLocked)
			<-step
			return nil
		})
		<-isLocked

		results := make(chan error, 1)
		for i := 0; i < 256; i++ {
			go func() {
				results <- executor.Execute(ctx, store.Mutex(333), func(ctx context.Context) error { return nil })
			}()
		}

		select {
		case <-results:
			t.Error("expected handlers to await lock")
		case <-time.After(time.Millisecond * 100):
		}
		step <- struct{}{}

		// see that mutex begins to be claimed by other executors.
		// only 10 because 256*CheckInInterval is a long time.
		done = make(chan struct{})
		go func() {
			for i := 0; i < 10; i++ {
				<-results
			}
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(time.Millisecond * 200):
			t.Error("expected other handlers to begin running")
		}
	})

	t.Run("handles lost lock", func(t *testing.T) {
		withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
			// set up a connection with idle_in_transaction_session_timeout set
			// low enough to boot our executor's lock
			conn, _ := db.Acquire(ctx)
			conn.Exec(ctx, "SET SESSION idle_in_transaction_session_timeout = '50ms'")

			executor := SynchronizedExecutor{
				DB:              conn,
				CheckinInterval: time.Millisecond * 250,
			}
			hasLock, lostLock := make(chan struct{}), make(chan struct{})
			actualErr := executor.Execute(ctx, store.Mutex(1), func(ctx context.Context) error {
				close(hasLock)
				<-ctx.Done()
				close(lostLock)
				return nil
			})

			if actualErr != nil {
				t.Errorf("expected Execute to lose lock, and signal the handler to shutdown gracefully. %v", actualErr)
			}
			if _, ok := <-hasLock; ok {
				t.Fatal("expected to acquire lock")
			}
			if _, ok := <-lostLock; ok {
				t.Fatal("expected to have recieved cancel signal")
			}
		})
	})
}
