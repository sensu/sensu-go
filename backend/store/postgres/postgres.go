package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/echlebek/migration"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sensu/sensu-go/util/retry"
)

// Open opens a new postgresql database for event storage. If the function
// returns nil error, then the database will be upgraded to the latest schema
// version, and will be ready to be passed to the NewEventStore function.
func OpenEvents(ctx context.Context, config *pgxpool.Config, retryForever bool) (*pgxpool.Pool, error) {
	backoff := retry.ExponentialBackoff{
		Ctx:                  ctx,
		MaxRetryAttempts:     3,
		InitialDelayInterval: time.Second,
		MaxDelayInterval:     time.Second * 5,
		Multiplier:           2,
	}
	if retryForever {
		backoff.MaxRetryAttempts = 0
	}
	var db *pgxpool.Pool
	err := backoff.Retry(func(retry int) (bool, error) {
		var err error
		if db, err = migration.Open(config, migrations); err != nil {
			err = fmt.Errorf("error migrating database to latest version: %v", err)
			logger.WithError(err).Error("error opening postgres store, retrying...")
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}
