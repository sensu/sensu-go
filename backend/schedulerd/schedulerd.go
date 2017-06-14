package schedulerd

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// Schedulerd handles scheduling check requests for each check's
// configured interval and publishing to the message bus.
type Schedulerd struct {
	Store      store.Store
	MessageBus messaging.MessageBus

	schedulers *schedulerCollection
	wg         *sync.WaitGroup
	errChan    chan error
	stopping   chan struct{}
}

// Start the Scheduler daemon.
func (s *Schedulerd) Start() error {
	if s.Store == nil {
		return errors.New("no store available")
	}

	if s.MessageBus == nil {
		return errors.New("no message bus found")
	}

	s.schedulers = newSchedulerCollection(s.MessageBus, s.Store)
	s.errChan = make(chan error, 1)

	// The reconciler and the watchers have to be a little coordinated. We start
	// the watcher first, so that we don't miss any checks that are created
	// during our initial reconciliation phase.
	s.wg = &sync.WaitGroup{}
	s.stopping = make(chan struct{})
	s.reconcile()
	s.startReconciler()
	return nil
}

// Stop the scheduler daemon.
func (s *Schedulerd) Stop() error {
	close(s.stopping)
	s.schedulers.stopAll()
	// let the event queue drain so that we don't panic inside the loop.
	// TODO(greg): get ride of this dependency.
	s.wg.Wait()
	close(s.errChan)
	return nil
}

// Status returns the health of the scheduler daemon.
func (s *Schedulerd) Status() error {
	return nil
}

// Err returns a channel on which to listen for terminal errors.
func (s *Schedulerd) Err() <-chan error {
	return s.errChan
}

func (s *Schedulerd) startReconciler() {
	s.wg.Add(1)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for {
			select {
			case <-s.stopping:
				ticker.Stop()
				s.wg.Done()
				return
			case <-ticker.C:
				s.reconcile()
			}
		}
	}()
}

func (s *Schedulerd) reconcile() error {
	checks, err := s.Store.GetCheckConfigs("")
	if err != nil {
		return err
	}

	for _, check := range checks {
		if err = s.schedulers.add(check); err != nil {
			return err
		}
	}

	return err
}

type schedulerCollection struct {
	items map[string]*CheckScheduler
	mutex *sync.Mutex

	stopped *atomic.Value
	wg      *sync.WaitGroup

	createScheduler func(check *types.CheckConfig) *CheckScheduler
}

func newSchedulerCollection(msgBus messaging.MessageBus, store store.Store) *schedulerCollection {
	wg := &sync.WaitGroup{}
	stopped := &atomic.Value{}
	stopped.Store(false)

	createScheduler := func(check *types.CheckConfig) *CheckScheduler {
		return &CheckScheduler{
			CheckConfig: check,
			MessageBus:  msgBus,
			Store:       store,
			wg:          wg,
		}
	}

	collection := &schedulerCollection{
		createScheduler: createScheduler,
		items:           map[string]*CheckScheduler{},
		mutex:           &sync.Mutex{},
		stopped:         stopped,
		wg:              wg,
	}

	return collection
}

func (collectionPtr *schedulerCollection) add(check *types.CheckConfig) error {
	// Guard against updates while the daemon is shutting down
	if collectionPtr.stopped.Load().(bool) {
		return nil
	}

	// Avoid competing updates
	collectionPtr.mutex.Lock()
	defer collectionPtr.mutex.Unlock()

	// Guard against creating a duplicate scheduler; schedulers are able to update
	// their internal state with any changes that occur to their associated check.
	if existing := collectionPtr.items[check.Name]; existing != nil {
		return nil
	}

	// Create new scheduler & start it
	scheduler := collectionPtr.createScheduler(check)
	if err := scheduler.Start(); err != nil {
		return err
	}

	// Register new check scheduler
	collectionPtr.items[check.Name] = scheduler
	return nil
}

func (collectionPtr *schedulerCollection) stopAll() {
	// Await any pending updates before shutting down
	collectionPtr.mutex.Lock()
	collectionPtr.stopped.Store(true)

	for n, scheduler := range collectionPtr.items {
		delete(collectionPtr.items, n)
		scheduler.Stop()
	}

	collectionPtr.wg.Wait()
}
