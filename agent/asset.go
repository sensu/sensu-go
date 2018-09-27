package agent

import (
	"path/filepath"
	"time"

	"github.com/sensu/sensu-go/asset"
	bolt "go.etcd.io/bbolt"
)

const (
	dbName = "assets.db"
)

// startAssetManager starts the agent's asset manager.
func (a *Agent) startAssetManager() (asset.Getter, error) {
	db, err := bolt.Open(filepath.Join(a.config.CacheDir, dbName), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	a.wg.Add(1)
	go func() {
		<-a.stopping
		if err := db.Close(); err != nil {
			logger.Debug(err)
		}
		a.wg.Done()
	}()
	boltDBGetter := asset.NewBoltDBGetter(
		db, a.config.CacheDir, nil, nil, nil)

	return asset.NewFilteredManager(boltDBGetter, a.entity), nil
}
