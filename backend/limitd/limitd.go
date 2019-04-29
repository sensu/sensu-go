package limitd

import (
	"context"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/limiter"
	"github.com/sensu/sensu-go/backend/store/etcd"
)

const (
	// componentName identifies LimitD as the component/daemon implemented in this package.
	componentName = "limitd"

	// interval is the interval, in seconds, that LimitD will update the entity count.
	interval = 300 * time.Second
)

// Limitd is the limit daemon.
type Limitd struct {
	interval      time.Duration
	client        *clientv3.Client
	ctx           context.Context
	cancel        context.CancelFunc
	errChan       chan error
	entityLimiter *limiter.EntityLimiter
}

// Option is a functional option.
type Option func(*Limitd) error

// Config configures Limitd.
type Config struct {
	Client        *clientv3.Client
	EntityLimiter *limiter.EntityLimiter
}

// New creates a new Limitd.
func New(c Config, opts ...Option) (*Limitd, error) {
	t := &Limitd{
		interval:      interval,
		client:        c.Client,
		errChan:       make(chan error, 1),
		entityLimiter: c.EntityLimiter,
	}
	t.ctx, t.cancel = context.WithCancel(context.Background())

	return t, nil
}

// Start the limit daemon.
func (t *Limitd) Start() error {
	if err := t.ctx.Err(); err != nil {
		return err
	}

	go t.start()

	return nil
}

// Stop the limit daemon.
func (t *Limitd) Stop() error {
	t.cancel()
	close(t.errChan)
	return nil
}

// Err returns a channel on which to listen for terminal errors.
func (t *Limitd) Err() <-chan error {
	return t.errChan
}

// Name returns the daemon name.
func (t *Limitd) Name() string {
	return componentName
}

// start starts a loop to periodically update entity count and history.
func (t *Limitd) start() {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()
	t.count()
	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			t.count()
		}
	}
}

// count collects the total number of entities and adds the entry
// to the entity limiter history.
func (t *Limitd) count() {
	entities, err := etcd.Count(t.ctx, t.client, etcd.GetEntitiesPath(t.ctx, ""))
	if err != nil {
		logger.WithError(err).Error("unable to retrieve entity count")
		return
	}
	logger.WithField("entities", entities).Debug("counted total number of entities")
	t.entityLimiter.AddCount(int(entities))
}
