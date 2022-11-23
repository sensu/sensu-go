package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestOPCQuerySyntax(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		tx, err := db.Begin(ctx)
		if err != nil {
			t.Error(err)
			return
		}
		defer tx.Rollback(ctx)
		queries := []string{
			getOperatorID,
			opcCheckInInsert,
			opcCheckInUpdate,
			opcCheckOut,
			opcGetNotifications,
			opcUpdateNotifications,
			opcReassignAbsentControllers,
			opcGetOperator,
			opcGetOperatorByID,
		}
		for i := range queries {
			_, err := tx.Prepare(ctx, fmt.Sprintf("%d", i), queries[i])
			if err != nil {
				t.Errorf("query %d: %s", i, err)
			}
		}
	})
}
