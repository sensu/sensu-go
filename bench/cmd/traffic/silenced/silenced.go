package silenced

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"
)

// 1txn per 100ms per worker.
// ~1 cache rebuild per worker per second.
// ~10 resolved events updating silences per second.
func ReadWriteConfig(ctx context.Context, r *rand.Rand, tx *sql.Tx) error {
	ns := r.Intn(16)
	op := r.Intn(1000)
	select {
	case <-ctx.Done():
		return nil
	case <-time.After(time.Millisecond * 100):
	}
	switch {
	case op > 900:
		// 1 in 10 are full index scans to rebuild cache
		// read all config for a namespace
		q, err := tx.QueryContext(ctx, `SELECT id, resource FROM configuration
		WHERE api_version = 'core/v2' AND type = 'Silenced' AND namespace = $1`, fmt.Sprintf("ns-%d", ns))
		if err != nil {
			return err
		}
		defer q.Close()
		for q.Next() {
			var id int64
			var resource string
			if err := q.Scan(&id, &resource); err != nil {
				return err
			}
		}
		if err := q.Close(); err != nil {
			return err
		}
		fallthrough
	default:
		// Otherwise be eventd being busy
		eventdQ, err := tx.QueryContext(ctx, `SELECT id, resource FROM configuration
		WHERE api_version = 'core/v2' AND type = 'Silenced' AND namespace = $1
		AND name in ($2, $3, $4)`,
			fmt.Sprintf("ns-%d", ns),
			fmt.Sprintf("silenced-3-%d", ns),
			fmt.Sprintf("silenced-5-%d", ns),
			fmt.Sprintf("silenced-8-%d", ns))
		if err != nil {
			return err
		}
		var someId int64
		defer eventdQ.Close()
		for eventdQ.Next() {
			var id int64
			var resource string
			if err := eventdQ.Scan(&id, &resource); err != nil {
				return err
			}
			if id != 0 {
				someId = id
			}
		}
		if err := eventdQ.Close(); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE configuration SET deleted_at = NOW() WHERE id = $1`, someId)
		return err
	}
}

// 1txn per 100ms per worker.
// ~1 cache rebuild per worker per second.
// ~10 resolved events updating silences per second.
func ReadWriteDiscrete(ctx context.Context, r *rand.Rand, tx *sql.Tx) error {
	ns := r.Intn(16)
	target := r.Intn(4995)
	op := r.Intn(1000)
	select {
	case <-ctx.Done():
		return nil
	case <-time.After(time.Millisecond * 100):
	}
	switch {
	case op > 900:
		// 1 in 10 are full index scans to rebuild cache
		// read all config for a namespace
		q, err := tx.QueryContext(ctx, `SELECT id, name FROM silenced
		WHERE namespace = $1`, fmt.Sprintf("ns-%d", ns))
		if err != nil {
			return err
		}
		defer q.Close()
		for q.Next() {
			var id int64
			var resource string
			if err := q.Scan(&id, &resource); err != nil {
				return err
			}
		}
		if err := q.Close(); err != nil {
			return err
		}
		fallthrough
	default:
		// Otherwise be eventd being busy
		eventdQ, err := tx.QueryContext(ctx, `SELECT id, name FROM silenced
		WHERE namespace = $1
		AND name in ($2, $3, $4)`,
			fmt.Sprintf("ns-%d", ns),
			fmt.Sprintf("silenced-%d-%d", target, ns),
			fmt.Sprintf("silenced-%d-%d", target+1, ns),
			fmt.Sprintf("silenced-%d-%d", target+2, ns))
		if err != nil {
			return err
		}
		var someId int64
		defer eventdQ.Close()
		for eventdQ.Next() {
			var id int64
			var resource string
			if err := eventdQ.Scan(&id, &resource); err != nil {
				return err
			}
			if id != 0 {
				someId = id
			}
		}
		if err := eventdQ.Close(); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `DELETE FROM silenced WHERE id = $1`, someId)
		return err
	}
}
