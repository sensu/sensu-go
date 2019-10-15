package schedulerd

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/prometheus/client_golang/prometheus"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/cache"
	"github.com/sensu/sensu-go/types"
)

var (
	intervalCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sensu_go_interval_schedulers",
			Help: "Number of active interval check schedulers on this backend",
		},
		[]string{"namespace"})

	cronCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sensu_go_cron_schedulers",
			Help: "Number of active cron check schedulers on this backend",
		},
		[]string{"namespace"})

	rrIntervalCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sensu_go_round_robin_interval_schedulers",
			Help: "Number of active round robin interval check schedulers on this backend.",
		},
		[]string{"namespace"})

	rrCronCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sensu_go_round_robin_cron_schedulers",
			Help: "Number of active round robin cron check schedulers on this backend.",
		},
		[]string{"namespace"})
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
	entityCache          *cache.Resource
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
func New(ctx context.Context, c Config, opts ...Option) (*Schedulerd, error) {
	s := &Schedulerd{
		store:       c.Store,
		queueGetter: c.QueueGetter,
		bus:         c.Bus,
		errChan:     make(chan error, 1),
		ringPool:    c.RingPool,
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	cache, err := cache.New(s.ctx, c.Client, &corev2.Entity{}, true)
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
