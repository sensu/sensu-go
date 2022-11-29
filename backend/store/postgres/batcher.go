package postgres

import (
	"context"
	"errors"
	"sync"
	"time"

	pgxv5 "github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"
)

var ErrBatchFull = errors.New("batch full")

func init() {
	_ = prometheus.Register(batchesProcessed)
	_ = prometheus.Register(batchErrors)
	_ = prometheus.Register(batchCapacity)
}

var (
	batchesProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sensu_go_postgres_batches",
			Help: "The total number of event batches sent to postgres",
		},
		[]string{"batchsize"},
	)
	batchErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sensu_go_postgres_batch_errors",
			Help: "The total number of errors encountered on batch writes",
		},
		[]string{},
	)
	batchCapacity = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sensu_go_postgres_free_batches",
			Help: "number of postgres batches ready to work",
		},
		[]string{"state"},
	)
)

var BatchTTL = 100 * time.Millisecond

const (
	// DefaultBatchBufferSize is the default size of the batch buffer.
	DefaultBatchBufferSize = 10000
)

// EventBatcher batches queries.
type EventBatcher struct {
	work chan *workItem
	db   DB
}

// DB is the storage interface for the batcher.
type DB interface {
	SendBatch(context.Context, *pgxv5.Batch) pgxv5.BatchResults
}

type BatchConfig struct {
	// Producers is the number of event producer goroutines. Should be set to
	// the number of eventd worker goroutines.
	Producers int

	// Consumers is the number of event consumer goroutines.
	Consumers int

	// BatchSize is the size of each batch sent to postgres.
	BatchSize int

	// BufferSize is the size of the event buffer that the consumers consume
	// form.
	BufferSize int

	// DB is the storage abstraction for the batcher. It is based on pgx's
	// batching facility.
	DB DB
}

// NewEventBatcher creates a new EventBatcher.
func NewEventBatcher(ctx context.Context, cfg BatchConfig) (*EventBatcher, error) {
	if cfg.Producers < 1 {
		cfg.Producers = 1
	}
	if cfg.Consumers < 1 {
		cfg.Consumers = 1
	}
	if cfg.BatchSize < 1 {
		cfg.BatchSize = 1
	}
	if cfg.BatchSize*cfg.Consumers > cfg.Producers {
		cfg.Consumers = cfg.Producers / cfg.BatchSize
	}
	e := &EventBatcher{
		work: make(chan *workItem, cfg.BufferSize),
		db:   cfg.DB,
	}
	workerWg := new(sync.WaitGroup)
	workerWg.Add(cfg.Consumers)
	for i := 0; i < cfg.Consumers; i++ {
		go func() {
			defer workerWg.Done()
			e.worker(ctx, cfg.BatchSize)
		}()
	}
	return e, nil
}

func itemEq(a, b EventArgs) bool {
	return a.Namespace == b.Namespace && a.Entity == b.Entity && a.Check == b.Check
}

func uniqueAppend(buffer []*workItem, work *workItem) (result []*workItem, ok bool) {
	for i := range buffer {
		if itemEq(buffer[i].args, work.args) {
			return buffer, false
		}
	}
	return append(buffer, work), true
}

func (e *EventBatcher) worker(ctx context.Context, batchSize int) {
	batch := &pgxv5.Batch{}
	works := make([]*workItem, 0, batchSize)
	ticker := time.NewTicker(BatchTTL)
	defer ticker.Stop()
	lastExec := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		case work := <-e.work:
			var ok bool
			works, ok = uniqueAppend(works, work)
			if !ok {
				logger.Error("duplicate simultaneous event! retrying...")
				go func() {
					e.work <- work
				}()
				continue
			}
			batch.Queue(UpdateEventQuery, work.args.Slice()...)
			if batch.Len() == batchSize {
				e.executeBatch(ctx, batch, works)

				// it's time for a new batch
				batch = &pgxv5.Batch{}
				works = works[:0]

				lastExec = time.Now()
			}
		case <-ticker.C:
			if batch.Len() > 0 && time.Since(lastExec) > BatchTTL {
				e.executeBatch(ctx, batch, works)

				// it's time for a new batch
				batch = &pgxv5.Batch{}
				works = works[:0]

				lastExec = time.Now()
			}
		}
	}
}

func fillErrors(work []*workItem, err error, seen map[int]struct{}) {
	for i, item := range work {
		if item.receipt.Err == nil {
			item.receipt.Err = err
		}
		if _, ok := seen[i]; !ok {
			close(item.receipt.C)
			seen[i] = struct{}{}
		}
	}
}

func (e *EventBatcher) executeBatch(ctx context.Context, batch *pgxv5.Batch, work []*workItem) {
	if batch.Len() == 0 {
		return
	}
	seen := make(map[int]struct{})
	results := e.db.SendBatch(ctx, batch)
	defer func() {
		if err := results.Close(); err != nil {
			fillErrors(work, err, seen)
		}
	}()
	for i := 0; i < batch.Len(); i++ {
		var row EventRow
		err := results.QueryRow().Scan(
			&row.HistoryTS,
			&row.HistoryStatus,
			&row.HistoryIndex,
			&row.LastOK,
			&row.Occurrences,
			&row.OccurrencesWatermark,
			&row.PreviousSerialized,
		)
		if err != nil {
			fillErrors(work, err, seen)
			return
		}
		receiver := work[i].receipt.C
		receiver <- row
		close(receiver)
		seen[i] = struct{}{}
	}
}

type EventArgs struct {
	Namespace  string
	Entity     string
	Check      string
	Status     int32
	LastOK     int64
	Serialized []byte
	Selectors  []byte
}

func (e *EventArgs) Slice() []interface{} {
	return []interface{}{
		e.Namespace,
		e.Entity,
		e.Check,
		e.Status,
		e.LastOK,
		e.Serialized,
		e.Selectors,
	}
}

type EventRow struct {
	HistoryIndex         int64
	LastOK               int64
	Occurrences          int64
	OccurrencesWatermark int64
	Index                int
	PreviousSerialized   []byte
	HistoryTS            []int64
	HistoryStatus        []int64
}

func (e *EventBatcher) Do(args EventArgs) (EventRow, error) {
	receipt := newWorkReceipt()
	e.work <- &workItem{
		receipt: receipt,
		args:    args,
	}
	row, ok := <-receipt.C
	if !ok && receipt.Err == nil {
		return row, errors.New("duplicate event dropped, data lost")
	}
	return row, receipt.Err
}

func newWorkReceipt() *workReceipt {
	return &workReceipt{
		C: make(chan EventRow, 1),
	}
}

type workReceipt struct {
	C   chan EventRow
	Err error
}

type workItem struct {
	args    EventArgs
	receipt *workReceipt
}
