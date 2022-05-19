package postgres

import (
	"time"

	"github.com/lib/pq"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/poll"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// Watcher is the postgres backed storev2.Watcher implementation
type Watcher struct {
	Builder WatchQueryBuilder
	Config  WatchConfig
}

func (w Watcher) Watch(req storev2.ResourceRequest) (<-chan storev2.WatchEvent, error) {
	if w.Config.OutputBufferSize <= 0 {
		w.Config.OutputBufferSize = 128
	}
	if w.Config.Interval <= 0 {
		w.Config.Interval = time.Second
	}
	if w.Config.TxnWindow <= 0 {
		w.Config.TxnWindow = 5 * time.Second
	}

	eventChan := make(chan storev2.WatchEvent, w.Config.OutputBufferSize)

	table := w.Builder.queryFor(req.StoreName, req.Namespace)

	p := poll.Poller{
		Interval:  w.Config.Interval,
		TxnWindow: w.Config.TxnWindow,
		Table:     table,
	}
	go p.Watch(req.Context, eventChan)
	return eventChan, nil
}

type WatchConfig struct {
	Interval         time.Duration
	TxnWindow        time.Duration
	OutputBufferSize int
}

// WatchQueryBuilder builds watcher queries for a specific postgres store backend
type WatchQueryBuilder interface {
	queryFor(storeName, namespace string) poll.Table
}

// recordStatus used by postgres stores implementing poll.Table
type recordStatus struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt pq.NullTime
}

// Row builds a poll.Row from a scanned row
func (rs recordStatus) Row(id string, resource corev3.Resource) poll.Row {
	row := poll.Row{
		Id:        id,
		Resource:  resource,
		CreatedAt: rs.CreatedAt,
		UpdatedAt: rs.UpdatedAt,
	}
	if rs.DeletedAt.Valid {
		row.DeletedAt = &rs.DeletedAt.Time
	}
	return row
}
