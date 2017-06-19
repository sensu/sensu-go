package schedulerd

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// ResourceSync interface for structs that fetch resources
type ResourceSync interface {
	Sync() error
}

// SyncronizeChecks fetches checks from the store and bubbles up results
type SyncronizeChecks struct {
	Store    store.CheckConfigStore
	OnUpdate func([]*types.CheckConfig)
}

// Sync fetches results from the store and passes them up w/ given handler
func (syncPtr *SyncronizeChecks) Sync() error {
	results, err := syncPtr.Store.GetCheckConfigs("")
	if err == nil {
		syncPtr.OnUpdate(results)
	}

	return err
}

// SyncronizeAssets fetches assets from the store and bubbles up results
type SyncronizeAssets struct {
	Store    store.AssetStore
	OnUpdate func([]*types.Asset)
}

// Sync fetches results from the store and passes them up w/ given handler
func (syncPtr *SyncronizeAssets) Sync() error {
	results, err := syncPtr.Store.GetAssets("")
	if err == nil {
		syncPtr.OnUpdate(results)
	}

	return err
}

// A SyncResourceScheduler schedules synchronization handlers to be run at a
// given interval.
type SyncResourceScheduler struct {
	Syncers []ResourceSync

	interval  *atomic.Value
	waitGroup *sync.WaitGroup
	stopping  chan struct{}
}

// NewSyncResourceScheduler instantiates new scheduler to sync resources
func NewSyncResourceScheduler(syncers []ResourceSync, interval int) *SyncResourceScheduler {
	scheduler := &SyncResourceScheduler{
		Syncers:   syncers,
		interval:  &atomic.Value{},
		waitGroup: &sync.WaitGroup{},
	}
	scheduler.SetInterval(interval)
	return scheduler
}

// SetInterval ...
func (recPtr *SyncResourceScheduler) SetInterval(i int) {
	recPtr.interval.Store(i)
}

// Start the scheduler
func (recPtr *SyncResourceScheduler) Start() {
	recPtr.Sync()

	recPtr.stopping = make(chan struct{})
	recPtr.waitGroup = &sync.WaitGroup{}
	recPtr.waitGroup.Add(1)

	go func() {
		interval := recPtr.interval.Load().(int)
		ticker := time.NewTicker(time.Duration(interval) * time.Second)

		for {
			select {
			case <-recPtr.stopping:
				ticker.Stop()
				recPtr.waitGroup.Done()
				return
			case <-ticker.C:
				recPtr.Sync()
			}
		}
	}()
}

// Sync executes all syncronizers
func (recPtr *SyncResourceScheduler) Sync() {
	for _, syncer := range recPtr.Syncers {
		go syncer.Sync()
	}
}

// Stop the scheduler
func (recPtr *SyncResourceScheduler) Stop() error {
	close(recPtr.stopping)
	recPtr.waitGroup.Wait()
	return nil
}
