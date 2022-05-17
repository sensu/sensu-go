package postgres

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/robfig/cron/v3"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sirupsen/logrus"
)

type Ring struct {
	db        *pgxpool.Pool
	namespace string
	name      string
	path      string
	logger    *logrus.Entry
	wg        sync.WaitGroup
	bus       *Bus
}

func (r *Ring) Close() error {
	return nil
}

func unPath(key string) (namespace, subscription string, err error) {
	parts := strings.Split(key, "/")
	if len(parts) < 5 {
		return "", "", errors.New("invalid ring key: " + key)
	}
	return parts[3], parts[4], nil
}

func NewRing(db *pgxpool.Pool, bus *Bus, path string) (*Ring, error) {
	ring := Ring{
		db:   db,
		path: path,
		bus:  bus,
	}
	var err error
	ring.namespace, ring.name, err = unPath(path)
	if err != nil {
		logger.WithError(err).Debug("error parsing ring path")
		return nil, err
	}
	ring.logger = logger.WithField("namespace", ring.namespace).WithField("subscription", ring.name)
	ring.logger.Info("initializing round-robin ring")
	if _, err := ring.db.Exec(context.Background(), insertRingQuery, ring.path); err != nil {
		return nil, err
	}
	ring.logger.Trace("ring ready for use")
	return &ring, nil
}

func (r *Ring) Subscribe(ctx context.Context, sub ringv2.Subscription) <-chan ringv2.Event {
	r.logger.Tracef("subscribing to %s", sub.Name)
	result := make(chan ringv2.Event, 1)
	row := r.db.QueryRow(ctx, insertRingSubscriberQuery, r.path, sub.Name)
	var inserted bool
	if err := row.Scan(&inserted); err != nil && err != pgx.ErrNoRows {
		result <- ringv2.Event{
			Type: ringv2.EventError,
			Err:  err,
		}
		logger.WithError(err).Error("error inserting ring subscriber")
		close(result)
		return result
	}
	r.logger.WithField("new subscription", !inserted).Tracef("subscribed to %s", sub.Name)
	r.wg.Add(2)
	go r.manage(ctx, sub)
	go r.produce(ctx, sub, result)
	return result
}

type ticker struct {
	C            <-chan time.Time
	Stop         func()
	NextDuration func() time.Duration
}

func (r *Ring) manage(ctx context.Context, sub ringv2.Subscription) {
	defer r.wg.Done()
	ticker := getTicker(ctx, sub)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			r.logger.Tracef("shutting down ring subscription: %s", sub.Name)
			return
		case <-ticker.C:
			r.logger.Tracef("updating ring subscription: %s", sub.Name)
			r.doManage(ctx, sub, ticker.NextDuration())
		}
	}
}

func (r *Ring) doManage(ctx context.Context, sub ringv2.Subscription, dur time.Duration) {
	logger := r.logger.WithField("check", sub.Name)
	logger.WithField("duration", dur.String()).Trace("attempting to advance the ring")
	tx, err := r.db.Begin(ctx)
	if err != nil {
		logger.WithError(err).Error("error incrementing round-robin ring")
		return
	}
	defer func() {
		if err == nil || err == pgx.ErrNoRows {
			if err = tx.Commit(ctx); err != nil {
				logger.WithError(err).Error("error committing ring update transaction")
			}
			return
		}
		logger.WithError(err).Error("ring database error")
		logger.Info("rolling back transaction")
		err := tx.Rollback(context.Background())
		if err != nil {
			logger.WithError(err).Error("error rolling back ring transaction")
			return
		}
		logger.Info("transaction rolled back")
	}()

	// Add jitter to the duration and subtract it from the duration. This
	// achieves two things. One, it increases the likelihood that the query
	// will not "undershoot", leaving an un-incremented ring for a given interval.
	// Two, it should lead to a better distribution of ring management between
	// backends.
	jitter := rand.Float64()
	dur = dur - time.Duration(float64(time.Millisecond)*jitter)

	logger.WithField("path", r.path).WithField("items", sub.Items).WithField("dur", dur.String()).Trace("query updateRingSubscribers")
	row := tx.QueryRow(ctx, updateRingSubscribersQuery, r.path, sub.Name, sub.Items-1, dur.String())
	var nextEntity string
	if err = row.Scan(&nextEntity); err != nil {
		if err == pgx.ErrNoRows {
			// expected case when another backend has updated the ring already,
			// or the ring is empty.
			logger.Trace("ring advanced by another backend")
			err = nil
			return
		}
		logger.WithError(err).Error("error updating ring subscribers")
		return
	}
	logger.WithField("next_entity", nextEntity).Trace("sending ring notification to postgres")
	if _, err = tx.Exec(ctx, notifyRingChannelQuery, ListenChannelName(r.namespace, sub.Name)); err != nil {
		return
	}
}

func getTicker(ctx context.Context, sub ringv2.Subscription) ticker {
	var ticker ticker
	if sub.IntervalSchedule > 0 {
		dur := time.Second * time.Duration(sub.IntervalSchedule)
		tkr := time.NewTicker(dur)
		ticker.C = tkr.C
		ticker.Stop = tkr.Stop
		ticker.NextDuration = func() time.Duration { return dur }
	} else {
		sched, err := cron.ParseStandard(sub.CronSchedule)
		if err != nil {
			logger.WithField("check", sub.Name).WithError(err).Error("invalid cron schedule!")
			ticker.NextDuration = func() time.Duration { return time.Minute }
			return ticker
		}
		ch := make(chan time.Time, 1)
		ticker.Stop = func() {}
		ticker.NextDuration = func() time.Duration {
			return time.Until(sched.Next(time.Now()))
		}
		ticker.C = ch
		go cronLoop(ctx, sched, ch)
	}
	return ticker
}

func cronLoop(ctx context.Context, sched cron.Schedule, ch chan time.Time) {
	timer := time.NewTimer(time.Until(sched.Next(time.Now())))
	for {
		select {
		case <-ctx.Done():
			close(ch)
			if !timer.Stop() {
				<-timer.C
			}
			return
		case ch <- <-timer.C:
			timer.Reset(time.Until(sched.Next(time.Now())))
		}
	}
}

func (r *Ring) produce(ctx context.Context, sub ringv2.Subscription, ch chan ringv2.Event) {
	logger := r.logger.WithField("check", sub.Name)
	logger.Trace("produce()")
	defer r.wg.Done()
	defer close(ch)
	notifications, err := r.bus.Subscribe(ctx, r.namespace, sub.Name)
	if err != nil {
		ch <- ringv2.Event{
			Type: ringv2.EventError,
			Err:  err,
		}
		logger.WithError(err).Error("error setting up postgres notification listener")
		return
	}
	r.logger.Trace("established a postgres listener")
	ch <- r.doProduce(ctx, sub)
	for {
		select {
		case <-ctx.Done():
			logger.Trace("context canceled")
			return
		case <-notifications:
			ch <- r.doProduce(ctx, sub)
		}
	}
}

func (r *Ring) doProduce(ctx context.Context, sub ringv2.Subscription) ringv2.Event {
	logger := r.logger.WithField("check", sub.Name)
	logger.Trace("doProduce()")
	rows, err := r.db.Query(ctx, getRingEntitiesQuery, r.path, sub.Name, sub.Items)
	if err != nil {
		return ringv2.Event{
			Type: ringv2.EventError,
			Err:  err,
		}
	}
	defer rows.Close()
	event := ringv2.Event{
		Type:   ringv2.EventTrigger,
		Values: make([]string, 0, sub.Items),
	}
	for rows.Next() {
		logger.Trace("rows.Next()")
		var entity string
		if err := rows.Scan(&entity); err != nil {
			return ringv2.Event{
				Type: ringv2.EventError,
				Err:  err,
			}
		}
		logger.WithField("entity", entity).Trace("got an entity")
		event.Values = append(event.Values, entity)
	}
	if len(event.Values) != 0 && len(event.Values) < sub.Items {
		more := make([]string, 0, sub.Items-len(event.Values))
		for i := len(event.Values); i < sub.Items; i++ {
			more = append(more, event.Values[i%len(event.Values)])
		}
		event.Values = append(event.Values, more...)
	}
	logger.Tracef("%#v\n", event)
	return event
}

func (r *Ring) Remove(ctx context.Context, value string) error {
	if ctx.Value(ringv2.DeleteEntityContextKey) != nil {
		_, err := r.db.Exec(ctx, deleteEntityQuery, r.namespace, value)
		return err
	}
	_, err := r.db.Exec(ctx, deleteRingEntityQuery, r.namespace, r.path, value)
	return err
}

func (r *Ring) Add(ctx context.Context, value string, keepalive int64) (err error) {
	r.logger.WithField("entity", value).WithField("keepalive", keepalive).Trace("ring.Add()")
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit(ctx)
			return
		}
		if txErr := tx.Rollback(context.Background()); txErr != nil {
			err = txErr
		}
	}()
	dur := time.Duration(keepalive) * time.Second
	if _, err := tx.Exec(ctx, insertEntityQuery, r.namespace, value, dur.String()); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, insertRingEntityQuery, r.namespace, value, r.path); err != nil {
		return err
	}
	return nil
}

func (r *Ring) IsEmpty(ctx context.Context) (bool, error) {
	r.logger.Trace("ring.IsEmpty()")
	row := r.db.QueryRow(ctx, getRingLengthQuery, r.path)
	var count int64
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count == 0, nil
}
