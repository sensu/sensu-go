package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sensu/sensu-go/backend/queue/testspec"
)

func TestQueueSpec(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		q := &Queue{
			db:           db,
			pollDuration: time.Millisecond * 50,
		}

		testspec.RunClientTestSuite(t, ctx, q)
	})
}
