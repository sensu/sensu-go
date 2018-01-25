package schedulerd

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// ResourceSync interface for structs that fetch resources
type ResourceSync interface {
	Sync(ctx context.Context) error
}

// SynchronizeChecks fetches checks from the store and bubbles up results
type SynchronizeChecks struct {
	Store    store.CheckConfigStore
	OnUpdate func([]*types.CheckConfig)
}

// Sync fetches results from the store and passes them up w/ given handler
func (syncPtr *SynchronizeChecks) Sync(ctx context.Context) error {
	results, err := syncPtr.Store.GetCheckConfigs(ctx)
	if err == nil {
		syncPtr.OnUpdate(results)
	}

	return err
}

// SynchronizeAssets fetches assets from the store and bubbles up results
type SynchronizeAssets struct {
	Store    store.AssetStore
	OnUpdate func([]*types.Asset)
}

// Sync fetches results from the store and passes them up w/ given handler
func (syncPtr *SynchronizeAssets) Sync(ctx context.Context) error {
	results, err := syncPtr.Store.GetAssets(ctx)
	if err == nil {
		syncPtr.OnUpdate(results)
	}

	return err
}

// SynchronizeHooks fetches hooks from the store and bubbles up results
type SynchronizeHooks struct {
	Store    store.HookConfigStore
	OnUpdate func([]*types.HookConfig)
}

// Sync fetches results from the store and passes them up w/ given handler
func (syncPtr *SynchronizeHooks) Sync(ctx context.Context) error {
	results, err := syncPtr.Store.GetHookConfigs(ctx)
	if err == nil {
		syncPtr.OnUpdate(results)
	}

	return err
}

// SynchronizeEntities fetches entities from the store and bubbles up results
type SynchronizeEntities struct {
	Store    store.EntityStore
	OnUpdate func([]*types.Entity)
}

// Sync fetches results from the store and passes them up w/ given handler
func (syncPtr *SynchronizeEntities) Sync(ctx context.Context) error {
	results, err := syncPtr.Store.GetEntities(ctx)
	if err == nil {
		syncPtr.OnUpdate(results)
	}

	return err
}

// SynchronizeStateScheduler schedules synchronization handlers to be run at a
// given interval.
type SynchronizeStateScheduler struct {
	synchronizers []ResourceSync

	interval  *atomic.Value
	waitGroup *sync.WaitGroup
	stopping  chan struct{}
}

// NewSynchronizeStateScheduler instantiates new scheduler to sync resources
func NewSynchronizeStateScheduler(interval uint, syncs ...ResourceSync) *SynchronizeStateScheduler {
	scheduler := &SynchronizeStateScheduler{
		synchronizers: syncs,
		waitGroup:     &sync.WaitGroup{},
		interval:      &atomic.Value{},
	}
	scheduler.SetInterval(interval)
	return scheduler
}

// SetInterval ...
func (recPtr *SynchronizeStateScheduler) SetInterval(i uint) {
	recPtr.interval.Store(i)
}

// Start the scheduler
func (recPtr *SynchronizeStateScheduler) Start(ctx context.Context) {
	recPtr.Sync(ctx)

	recPtr.stopping = make(chan struct{})
	recPtr.waitGroup = &sync.WaitGroup{}
	recPtr.waitGroup.Add(1)

	go func() {
		interval := recPtr.interval.Load().(uint)
		ticker := time.NewTicker(time.Duration(interval) * time.Second)

		for {
			select {
			case <-recPtr.stopping:
				ticker.Stop()
				recPtr.waitGroup.Done()
				return
			case <-ticker.C:
				recPtr.Sync(ctx)
			}
		}
	}()
}

// Sync executes all synchronizers
func (recPtr *SynchronizeStateScheduler) Sync(ctx context.Context) {
	for _, syncer := range recPtr.synchronizers {
		syncer := syncer
		go func() {
			if err := syncer.Sync(ctx); err != nil {
				logger.Error(err)
			}
		}()
	}
}

// Stop the scheduler
func (recPtr *SynchronizeStateScheduler) Stop() error {
	close(recPtr.stopping)
	recPtr.waitGroup.Wait()
	return nil
}
