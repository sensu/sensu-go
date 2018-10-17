package asset

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sensu/sensu-go/types"
	bolt "go.etcd.io/bbolt"
)

const (
	dbName = "assets.db"
)

// Manager ...
type Manager struct {
	cacheDir string
	entity   *types.Entity
	stopping chan struct{}
	wg       *sync.WaitGroup
}

// NewManager ...
func NewManager(cacheDir string, entity *types.Entity, stopping chan struct{}, wg *sync.WaitGroup) *Manager {
	return &Manager{
		cacheDir: cacheDir,
		entity:   entity,
		stopping: stopping,
		wg:       wg,
	}
}

// StartAssetManager starts the asset manager for a backend or agent.
func (m *Manager) StartAssetManager() (Getter, error) {
	// create agent cache directory if it doesn't already exist
	if err := os.MkdirAll(m.cacheDir, 0755); err != nil {
		return nil, err
	}
	db, err := bolt.Open(filepath.Join(m.cacheDir, dbName), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	m.wg.Add(1)
	go func() {
		<-m.stopping
		if err := db.Close(); err != nil {
			logger.Debug(err)
		}
		m.wg.Done()
	}()
	boltDBGetter := NewBoltDBGetter(
		db, m.cacheDir, nil, nil, nil)

	return NewFilteredManager(boltDBGetter, m.entity), nil
}
