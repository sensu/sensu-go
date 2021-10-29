package schedulerd

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	cachev2 "github.com/sensu/sensu-go/backend/store/cache/v2"
	"github.com/sensu/sensu-go/types"
	clientv3 "go.etcd.io/etcd/client/v3"
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
	store                  store.Store
	queueGetter            types.QueueGetter
	bus                    messaging.MessageBus
	checkWatcher           *CheckWatcher
	adhocRequestExecutor   *AdhocRequestExecutor
	errChan                chan error
	ringPool               *ringv2.RingPool
	entityCache            *cachev2.Resource
	secretsProviderManager *secrets.ProviderManager
}

// Option is a functional option.
type Option func(*Schedulerd) error

// Config configures Schedulerd.
type Config struct {
	Store                  store.Store
	QueueGetter            types.QueueGetter
	RingPool               *ringv2.RingPool
	Bus                    messaging.MessageBus
	EntityCache            *cachev2.Resource
	Client                 *clientv3.Client
	SecretsProviderManager *secrets.ProviderManager
}

// New creates a new Schedulerd.
func New(c Config, opts ...Option) (*Schedulerd, error) {
	s := &Schedulerd{
		store:                  c.Store,
		queueGetter:            c.QueueGetter,
		bus:                    c.Bus,
		errChan:                make(chan error, 1),
		ringPool:               c.RingPool,
		secretsProviderManager: c.SecretsProviderManager,
	}
	s.entityCache = c.EntityCache
	s.checkWatcher = NewCheckWatcher(c.Bus, c.Store, c.RingPool, c.EntityCache, s.secretsProviderManager)

	for _, o := range opts {
		if err := o(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Start the Scheduler daemon.
func (s *Schedulerd) Start(ctx context.Context) error {
	s.adhocRequestExecutor = NewAdhocRequestExecutor(ctx, s.store, s.queueGetter.GetQueue(adhocQueueName), s.bus, s.entityCache, s.secretsProviderManager)

	_ = prometheus.Register(intervalCounter)
	_ = prometheus.Register(cronCounter)
	_ = prometheus.Register(rrIntervalCounter)
	_ = prometheus.Register(rrCronCounter)

	return s.checkWatcher.Start(ctx)
}

// Stop the scheduler daemon.
func (s *Schedulerd) Stop() error {
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
