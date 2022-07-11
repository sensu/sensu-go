package postgres

import (
	"sync"

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
