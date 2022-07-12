package postgres

import (
	"database/sql"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sensu/sensu-go/backend/poll"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// watchStoreOverrides contains per-store overrides to the default watcher query builder.
var watchStoreOverridesMu sync.Mutex
var watchStoreOverrides map[string]watchStoreFactory = make(map[string]watchStoreFactory)

type watchStoreFactory func(storev2.ResourceRequest, *pgxpool.Pool) (poll.Table, error)

func registerWatchStoreOverride(storeName string, factory watchStoreFactory) {
	watchStoreOverridesMu.Lock()
	defer watchStoreOverridesMu.Unlock()
	watchStoreOverrides[storeName] = factory
}

func getWatchStoreOverride(storeName string) (factory watchStoreFactory, ok bool) {
	watchStoreOverridesMu.Lock()
	defer watchStoreOverridesMu.Unlock()
	factory, ok = watchStoreOverrides[storeName]
	return
}

// recordStatus used by postgres stores implementing poll.Table
type recordStatus struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime
}

// Row builds a poll.Row from a scanned row
func (rs recordStatus) Row(id string, resource storev2.Wrapper) poll.Row {
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
