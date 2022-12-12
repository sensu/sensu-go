package schedulerd

import (
	"context"
	"fmt"
	"strings"

	time "github.com/echlebek/timeproxy"
	"github.com/prometheus/client_golang/prometheus"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/secrets"
	cachev2 "github.com/sensu/sensu-go/backend/store/cache/v2"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
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
	schedRefreshDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "sensu_go_schedulerd_refresh_duration",
			Help:    "Duration of schedulerd's refresh opration in seconds",
			Buckets: []float64{0.005, 0.01, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
	)
)

// Schedulerd handles scheduling check requests for each check's
// configured interval and publishing to the message bus.
type Schedulerd struct {
	refreshInterval        time.Duration
	store                  storev2.Interface
	bus                    messaging.MessageBus
	ctx                    context.Context
	cancel                 context.CancelFunc
	errChan                chan error
	ringPool               *ringv2.RingPool
	entityCache            EntityCache
	secretsProviderManager *secrets.ProviderManager

	checks     namespacedChecks
	schedulers map[string]Scheduler
}

// Config configures Schedulerd.
type Config struct {
	Store                  storev2.Interface
	RingPool               *ringv2.RingPool
	Bus                    messaging.MessageBus
	SecretsProviderManager *secrets.ProviderManager
	RefreshInterval        time.Duration
}

// New creates a new Schedulerd.
func New(ctx context.Context, c Config) (*Schedulerd, error) {
	s := &Schedulerd{
		refreshInterval:        c.RefreshInterval,
		store:                  c.Store,
		bus:                    c.Bus,
		errChan:                make(chan error, 1),
		ringPool:               c.RingPool,
		secretsProviderManager: c.SecretsProviderManager,

		checks:     make(namespacedChecks),
		schedulers: make(map[string]Scheduler),
	}
	if s.refreshInterval <= 0 {
		s.refreshInterval = time.Second * 5
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	cache, err := cachev2.New[*corev3.EntityConfig](s.ctx, c.Store, true)
	if err != nil {
		return nil, err
	}
	s.entityCache = cache

	return s, nil
}

// Start the Scheduler daemon.
func (s *Schedulerd) Start() error {
	_ = prometheus.Register(intervalCounter)
	_ = prometheus.Register(cronCounter)
	_ = prometheus.Register(rrIntervalCounter)
	_ = prometheus.Register(rrCronCounter)
	_ = prometheus.Register(schedRefreshDuration)
	return s.start()
}

// start initializes schedulerd and begins polling for scheduling state changes
func (s *Schedulerd) start() error {
	if err := s.refresh(); err != nil {
		return err
	}
	go func() {
		tick := time.NewTicker(s.refreshInterval)
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-tick.C:
				if err := s.refresh(); err != nil {
					logger.WithError(err).Error("error refreshing scheduler")
				}
			}
		}
	}()
	return nil
}

// refresh the desired scheduler state
func (s *Schedulerd) refresh() error {
	timer := prometheus.NewTimer(schedRefreshDuration)
	defer timer.ObserveDuration()
	checkStore := storev2.Of[*corev2.CheckConfig](s.store)
	next, err := checkStore.List(s.ctx, corev2.ObjectMeta{}, nil)
	if err != nil {
		return err
	}
	added, changed, removed := s.checks.Update(next)

	checksAdded := make([]string, len(added))
	checksChanged := make([]string, len(changed))
	checksRemoved := make([]string, len(removed))
	for i, check := range added {
		checksAdded[i] = fmt.Sprintf("%s/%s", check.Namespace, check.Name)
		// Guard against updates while the daemon is shutting down
		if err := s.ctx.Err(); err != nil {
			return err
		}

		// Guard against creating a duplicate scheduler; schedulers are able to update
		// their internal state with any changes that occur to their associated check.
		// likely obsolete now that we've migrated off etcd watchers
		key := concatUniqueKey(check.Name, check.Namespace)
		if existing := s.schedulers[key]; existing != nil {
			if existing.Type() == GetSchedulerType(check) {
				logger.Error("scheduler already exists")
				return nil
			}
			if err := existing.Stop(); err != nil {
				return err
			}
		}

		var scheduler Scheduler

		switch GetSchedulerType(check) {
		case IntervalType:
			scheduler = NewIntervalScheduler(s.ctx, check, s.makeExecutor(check.Namespace))
		case CronType:
			scheduler = NewCronScheduler(s.ctx, check, s.makeExecutor(check.Namespace))
		case RoundRobinIntervalType:
			scheduler = NewRoundRobinIntervalScheduler(s.ctx, check, s.makeExecutor(check.Namespace), s.ringPool, s.entityCache)
		case RoundRobinCronType:
			scheduler = NewRoundRobinCronScheduler(s.ctx, check, s.makeExecutor(check.Namespace), s.ringPool, s.entityCache)
		default:
			logger.Error("bad scheduler type, falling back to interval scheduler")
			scheduler = NewIntervalScheduler(s.ctx, check, s.makeExecutor(check.Namespace))
		}

		// Start scheduling check
		scheduler.Start()

		// Register new check scheduler
		s.schedulers[key] = scheduler
	}
	if len(checksAdded) > 0 {
		logger.WithField("added", checksAdded).Info("added new checks to schedule")
	}

	for i, check := range changed {
		checksChanged[i] = fmt.Sprintf("%s/%s", check.Namespace, check.Name)
		key := concatUniqueKey(check.Name, check.Namespace)
		s.schedulers[key].Interrupt(check)
	}
	if len(checksChanged) > 0 {
		logger.WithField("changed", checksChanged).Info("updated schedule with new check configuration")
	}

	for i, check := range removed {
		checksRemoved[i] = fmt.Sprintf("%s/%s", check.Namespace, check.Name)
		key := concatUniqueKey(check.Name, check.Namespace)
		if err := s.schedulers[key].Stop(); err != nil {
			logger.WithError(err).Error("unexpected error stopping scheduler")
		}
		delete(s.schedulers, key)
	}
	if len(checksRemoved) > 0 {
		logger.WithField("removed", checksRemoved).Info("removed checks from schedule")
	}
	return nil

}

func (s *Schedulerd) makeExecutor(namespace string) *CheckExecutor {
	return NewCheckExecutor(s.bus, namespace, s.store, s.entityCache, s.secretsProviderManager)
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

func concatUniqueKey(args ...string) string {
	return strings.Join(args, "-")
}
