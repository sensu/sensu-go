package asset

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	corev2 "github.com/sensu/core/v2"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/time/rate"
)

const (
	dbName = "assets.db"
)

// Manager ...
type Manager struct {
	cacheDir      string
	entity        *corev2.Entity
	wg            *sync.WaitGroup
	trustedCAFile string
}

// NewManager ...
func NewManager(cacheDir, trustedCAFile string, entity *corev2.Entity, wg *sync.WaitGroup) *Manager {
	return &Manager{
		cacheDir:      cacheDir,
		entity:        entity,
		wg:            wg,
		trustedCAFile: trustedCAFile,
	}
}

// StartAssetManager starts the asset manager for a backend or agent.
func (m *Manager) StartAssetManager(ctx context.Context, limiter *rate.Limiter) (Getter, error) {
	// create agent cache directory if it doesn't already exist
	if err := os.MkdirAll(m.cacheDir, 0755); err != nil {
		return nil, err
	}

	logger.WithField("cache", m.cacheDir).Debug("initializing cache directory")
	db, err := bolt.Open(filepath.Join(m.cacheDir, dbName), 0600, &bolt.Options{Timeout: 60 * time.Second})
	if err != nil {
		return nil, err
	}
	logger.WithField("cache", m.cacheDir).Debug("done initializing cache directory")

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		<-ctx.Done()
		if err := db.Close(); err != nil {
			logger.Debug(err)
		}
	}()
	boltDBGetter := NewBoltDBGetter(
		db, m.cacheDir, m.trustedCAFile, nil, nil, nil, limiter)

	return NewFilteredManager(boltDBGetter, m.entity), nil
}
