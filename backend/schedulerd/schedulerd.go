package schedulerd

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/cache"
	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/types"
)

// Schedulerd handles scheduling check requests for each check's
// configured interval and publishing to the message bus.
type Schedulerd struct {
	store                store.Store
	queueGetter          types.QueueGetter
	bus                  messaging.MessageBus
	checkWatcher         *CheckWatcher
	adhocRequestExecutor *AdhocRequestExecutor
	ctx                  context.Context
	cancel               context.CancelFunc
	errChan              chan error
	ringPool             *ringv2.Pool
	entityCache          *cache.ResourceCacher
}

// Option is a functional option.
type Option func(*Schedulerd) error

// Config configures Schedulerd.
type Config struct {
	Store       store.Store
	QueueGetter types.QueueGetter
	RingPool    *ringv2.Pool
	Bus         messaging.MessageBus
	Client      *clientv3.Client
}

// New creates a new Schedulerd.
func New(c Config, opts ...Option) (*Schedulerd, error) {
	s := &Schedulerd{
		store:       c.Store,
		queueGetter: c.QueueGetter,
		bus:         c.Bus,
		errChan:     make(chan error, 1),
		ringPool:    c.RingPool,
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	cache, err := cache.New(s.ctx, c.Client, etcd.GetEntitiesPath, &corev2.Entity{})
	if err != nil {
		return nil, err
	}
	s.entityCache = cache
	s.checkWatcher = NewCheckWatcher(s.ctx, c.Bus, c.Store, c.RingPool, cache)
	s.adhocRequestExecutor = NewAdhocRequestExecutor(s.ctx, s.store, s.queueGetter.GetQueue(adhocQueueName), s.bus, s.entityCache)

	for _, o := range opts {
		if err := o(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Start the Scheduler daemon.
func (s *Schedulerd) Start() error {
	return s.checkWatcher.Start()
}

// Stop the scheduler daemon.
func (s *Schedulerd) Stop() error {
	s.cancel()
	close(s.errChan)
	return nil
}

// Err returns a channel on which to listen for terminal errors.
func (s *Schedulerd) Err() <-chan error {
	return s.errChan
}

// Name returns the daemon name
func (s *Schedulerd) Name() string {
	return "schedulerd"
}

func resourcesToEntities(resources []corev2.Resource) []*corev2.Entity {
	entities := make([]*corev2.Entity, len(resources))
	for i, resource := range resources {
		entities[i] = resource.(*corev2.Entity)
	}
	return entities
}
