package schedulerd

import (
	"sync"
	"time"

	"github.com/coreos/etcd/store"
	"github.com/sensu/sensu-go/types"
)

// ResourceSync interface for structs that fetch resources
type ResourceSync interface {
	Sync() error
}

// SyncronizeChecks fetches checks from the store and bubbles up results
type SyncronizeChecks struct {
	Store    store.CheckConfigStore
	OnUpdate func(*[]types.Check)
}

// Sync fetches results from the store and passes them up w/ given handler
func (syncPtr *CheckSyncer) Sync() error {
	results, err = syncPtr.Store.GetCheck("")
	if err == nil {
		syncPtr.OnUpdate(results)
	}

	return err
}

// SyncronizeAssets fetches assets from the store and bubbles up results
type SyncronizeAssets struct {
	Store    store.AssetStore
	OnUpdate func(*[]types.Asset)
}

// Sync fetches results from the store and passes them up w/ given handler
func (syncPtr *AssetSyncer) Sync() error {
	results, err = syncPtr.Store.GetAsset("")
	if err == nil {
		syncPtr.OnUpdate(results)
	}

	return nil
}

// A SyncResourceScheduler schedules synchronization handlers to be run at a
// given interval.
type SyncResourceScheduler struct {
	Syncers  []ResourceSync
	Interval uint64

	waitGroup *sync.WaitGroup
	stopping  chan struct{}
}

// Start the scheduler
func (recPtr *CheckReconciler) Start() {
	recPtr.Sync()

	recPtr.stopping = make(chan struct{})
	recPtr.waitGroup.Add(1)

	go func() {
		ticker := time.NewTicker(recPtr.Interval * time.Second)
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
func (recPtr *CheckReconciler) Sync() {
	for _, syncer := range recPtr.Syncers {
		go syncer.Sync()
	}
}

// Stop the scheduler
func (recPtr *CheckReconciler) Stop() error {
	close(recPtr.stopping)
	recPtr.waitGroup.Wait()
}
