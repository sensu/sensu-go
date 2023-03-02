package postgres

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/util/retry"
)

// SynchronizedExecutor functions as a postgresql advisory lock based mutex.
type SynchronizedExecutor struct {
	DB DBI

	CheckinInterval time.Duration
}

// Execute blocks until the Mutex can be locked, executes the handler function,
// then unlocks the Mutex. Implemented with postgresql advisory locks.
func (se *SynchronizedExecutor) Execute(ctx context.Context, mux store.Mutex, handler store.MutexHandler) error {
	backoff := retry.ExponentialBackoff{
		Ctx:                  ctx,
		InitialDelayInterval: se.CheckinInterval / 4,
		MaxDelayInterval:     se.CheckinInterval,
	}
	return backoff.Retry(func(_ int) (bool, error) {
		tx, err := se.tryLock(ctx, mux)
		if err != nil {
			logger.WithError(err).Errorf("error trying to lock mutex: %x", mux)
			return false, nil
		}
		if tx == nil {
			return false, nil
		}
		defer func() {
			if err := tx.Commit(ctx); err != nil {
				if !errors.Is(err, pgx.ErrTxClosed) && !errors.Is(err, ctx.Err()) {
					logger.WithError(err).
						Errorf("unexpected error committing transaction holding mutex lock. %x", mux)
				}
			}
		}()
		return true, se.handle(ctx, tx, mux, handler)
	})
}

func (se *SynchronizedExecutor) handle(ctx context.Context, conn DBI, mux store.Mutex, handler store.MutexHandler) error {
	handleCtx, cancel := context.WithCancel(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := se.checkin(handleCtx, conn, mux); err != nil {
			if errors.Is(err, handleCtx.Err()) {
				return
			}
			logger.WithError(err).
				Errorf("unexpected error holding mutex lock. considering mutex lost. %x", mux)
		}
		cancel()
	}()

	err := handler(handleCtx)
	cancel()
	wg.Wait()
	return err
}

func (se *SynchronizedExecutor) tryLock(ctx context.Context, mux store.Mutex) (pgx.Tx, error) {
	tx, err := se.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	locked, err := lockMux(ctx, tx, mux)
	if !locked || err != nil {
		_ = tx.Commit(ctx)
		return nil, err
	}

	return tx, err
}

// checkin periodically (re)locks the advisory lock in order to ensure that the
// txn is not booted for being idle. Postgres guarantees that lock requests for
// a lock already held by that transaction will always succeed.
func (se *SynchronizedExecutor) checkin(ctx context.Context, conn DBI, mux store.Mutex) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(se.CheckinInterval):
		}

		if _, err := lockMux(ctx, conn, mux); err != nil {
			return err
		}
	}
}

func lockMux(ctx context.Context, db DBI, mux store.Mutex) (bool, error) {
	row := db.QueryRow(ctx, "SELECT pg_try_advisory_xact_lock($1) as locked;", mux)
	var locked bool
	err := row.Scan(&locked)
	return locked, err
}
