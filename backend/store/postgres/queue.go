package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sensu/sensu-go/backend/queue"
)

type Queue struct {
	db           DBI
	pollDuration time.Duration
}

func NewQueue(db DBI) *Queue {
	return &Queue{
		db:           db,
		pollDuration: time.Second * 5,
	}
}

// Enqueue a queue item
func (q *Queue) Enqueue(ctx context.Context, item queue.Item) error {
	_, err := q.db.Exec(ctx, queueEnqueue, item.Queue, item.Value)
	return err
}

// Reserve reserves a queue item.
// When the queue is empty Reserve will block and poll for new items until either
// an item is found, or the context is cancelled.
//
// When Reserve returns a Reservation, the caller MUST Ack or Nack that Reservation.
// Otherwise a transaction + connection could be leaked depending on the session
// settings (i.e. idle_in_transaction_session_timeout.)
func (q *Queue) Reserve(ctx context.Context, queueName string) (queue.Reservation, error) {
	first := true
	for {
		if !first {
			select {
			case <-time.After(q.pollDuration):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		first = false

		tx, err := q.db.Begin(ctx)
		if err != nil {
			return nil, err
		}
		var item queue.Item

		row := tx.QueryRow(ctx, queueReserveItem, queueName)
		if err := row.Scan(&item.ID, &item.Queue, &item.Value); err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				return nil, fmt.Errorf("unexpected rollback error: %w", rollbackErr)
			}
			if err == pgx.ErrNoRows {
				continue
			}
			return nil, fmt.Errorf("error reserving queue item: %w", err)
		}
		return &queueReservation{
			tx:   tx,
			item: item,
		}, nil

	}
}

type queueReservation struct {
	tx   pgx.Tx
	item queue.Item
}

func (l *queueReservation) Item() queue.Item {
	return l.item
}

func (l *queueReservation) Ack(ctx context.Context) error {
	if _, err := l.tx.Exec(ctx, queueDeleteItem, l.item.ID); err != nil {
		defer func() { _ = l.Nack(ctx) }()
		return err
	}
	return l.tx.Commit(ctx)
}

func (l *queueReservation) Nack(ctx context.Context) error {
	return l.tx.Rollback(ctx)
}
