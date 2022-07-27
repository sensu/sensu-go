package schedulerd

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/secrets"
	cachev2 "github.com/sensu/sensu-go/backend/store/cache/v2"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/types"
	"go.etcd.io/etcd/client/v3"
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
	store                  storev2.Interface
	queueGetter            types.QueueGetter
	bus                    messaging.MessageBus
	checkWatcher           *CheckWatcher
	adhocRequestExecutor   *AdhocRequestExecutor
	ctx                    context.Context
	cancel                 context.CancelFunc
	errChan                chan error
	ringPool               *ringv2.RingPool
	entityCache            *cachev2.Resource
	secretsProviderManager *secrets.ProviderManager
}

// Option is a functional option.
type Option func(*Schedulerd) error

// Config configures Schedulerd.
type Config struct {
	Store                  storev2.Interface
	QueueGetter            types.QueueGetter
	RingPool               *ringv2.RingPool
	Bus                    messaging.MessageBus
	Client                 *clientv3.Client
	SecretsProviderManager *secrets.ProviderManager
}

// New creates a new Schedulerd.
func New(ctx context.Context, c Config, opts ...Option) (*Schedulerd, error) {
	s := &Schedulerd{
		store:                  c.Store,
		queueGetter:            c.QueueGetter,
		bus:                    c.Bus,
		errChan:                make(chan error, 1),
		ringPool:               c.RingPool,
		secretsProviderManager: c.SecretsProviderManager,
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	cache, err := cachev2.New(s.ctx, c.Store, &corev3.EntityConfig{}, true)
	if err != nil {
		return nil, err
	}
	s.entityCache = cache
	s.checkWatcher = NewCheckWatcher(s.ctx, c.Bus, c.Store, c.RingPool, cache, s.secretsProviderManager)
	s.adhocRequestExecutor = NewAdhocRequestExecutor(s.ctx, s.store, s.queueGetter.GetQueue(adhocQueueName), s.bus, s.entityCache, s.secretsProviderManager)

	for _, o := range opts {
		if err := o(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Start the Scheduler daemon.
func (s *Schedulerd) Start() error {
	_ = prometheus.Register(intervalCounter)
	_ = prometheus.Register(cronCounter)
	_ = prometheus.Register(rrIntervalCounter)
	_ = prometheus.Register(rrCronCounter)
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
