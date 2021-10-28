// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/pipeline"
	"github.com/sensu/sensu-go/backend/store"
	metricspkg "github.com/sensu/sensu-go/metrics"
)

const (
	// MessageHandlerDuration is the name of the prometheus summary vec used to
	// track average latencies of pipelined message handling.
	MessageHandlerDuration = "sensu_go_pipelined_message_handler_duration"

	// HasPipelinesLabelName is the name of a label which describes whether or
	// not the metric being recorded is for an event with pipelines.
	HasPipelinesLabelName = "has_pipelines"
)

var (
	defaultStoreTimeout = time.Minute

	messageHandlerDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       MessageHandlerDuration,
			Help:       "pipelined message handler latency distribution",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{metricspkg.StatusLabelName, HasPipelinesLabelName},
	)
)

// Pipelined handles incoming Sensu events and puts them through a
// Sensu event pipeline, i.e. filter -> mutator -> handler. The Sensu
// handler configuration determines which Sensu filters and mutator
// are used.
type Pipelined struct {
	stopping     chan struct{}
	running      *atomic.Value
	wg           *sync.WaitGroup
	errChan      chan error
	eventChan    chan interface{}
	subscription messaging.Subscription
	bus          messaging.MessageBus
	workerCount  int
	store        store.Store
	storeTimeout time.Duration
	adapters     []pipeline.Adapter
}

// Config configures a Pipelined.
type Config struct {
	Bus          messaging.MessageBus
	BufferSize   int
	Store        store.Store
	StoreTimeout time.Duration
	WorkerCount  int
}

// Option is a functional option used to configure Pipelined.
type Option func(*Pipelined) error

// PipelineGetter defines an interface for any structures which can return a
// slice of Pipeline resource references.
type PipelineGetter interface {
	GetPipelines() []*corev2.ResourceReference
	LogFields(bool) map[string]interface{}
}

// New creates a new Pipelined with supplied Options applied.
func New(c Config, options ...Option) (*Pipelined, error) {
	if c.BufferSize == 0 {
		logger.Warn("BufferSize not configured")
		c.BufferSize = 1
	}
	if c.WorkerCount == 0 {
		logger.Warn("WorkerCount not configured")
		c.WorkerCount = 1
	}
	if c.StoreTimeout == 0 {
		logger.Warn("StoreTimeout not configured")
		c.StoreTimeout = defaultStoreTimeout
	}

	p := &Pipelined{
		bus:          c.Bus,
		stopping:     make(chan struct{}, 1),
		running:      &atomic.Value{},
		wg:           &sync.WaitGroup{},
		errChan:      make(chan error, 1),
		eventChan:    make(chan interface{}, c.BufferSize),
		workerCount:  c.WorkerCount,
		store:        c.Store,
		storeTimeout: c.StoreTimeout,
	}
	for _, o := range options {
		if err := o(p); err != nil {
			return nil, err
		}
	}

	// Initialize labels & register metric families with Prometheus
	messageHandlerDuration.WithLabelValues(metricspkg.StatusLabelSuccess, "0")
	messageHandlerDuration.WithLabelValues(metricspkg.StatusLabelSuccess, "1")
	messageHandlerDuration.WithLabelValues(metricspkg.StatusLabelError, "0")
	messageHandlerDuration.WithLabelValues(metricspkg.StatusLabelError, "1")

	if err := prometheus.Register(messageHandlerDuration); err != nil {
		return nil, metricspkg.FormatRegistrationErr(MessageHandlerDuration, err)
	}

	return p, nil
}

// Receiver returns the event channel for pipelined.
func (p *Pipelined) Receiver() chan<- interface{} {
	return p.eventChan
}

// Start pipelined, subscribing to the "event" message bus topic to
// pass Sensu events to the pipelines for handling (goroutines).
func (p *Pipelined) Start() error {
	sub, err := p.bus.Subscribe(messaging.TopicEvent, "pipelined", p)
	if err != nil {
		return err
	}
	p.subscription = sub

	p.createWorkers(p.workerCount, p.eventChan)

	return nil
}

// Stop pipelined.
func (p *Pipelined) Stop() error {
	p.running.Store(false)
	close(p.stopping)
	p.wg.Wait()
	close(p.errChan)
	err := p.subscription.Cancel()
	close(p.eventChan)

	return err
}

// Err returns a channel to listen for terminal errors on.
func (p *Pipelined) Err() <-chan error {
	return p.errChan
}

// Name returns the daemon name
func (p *Pipelined) Name() string {
	return "pipelined"
}

func (p *Pipelined) AddAdapter(adapter pipeline.Adapter) {
	p.adapters = append(p.adapters, adapter)
}

// createWorkers creates several goroutines, responsible for pulling
// Sensu events from a channel (bound to message bus "event" topic)
// and passing them to their referenced pipelines.
func (p *Pipelined) createWorkers(count int, channel chan interface{}) {
	for i := 1; i <= count; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case <-p.stopping:
					return
				case msg := <-channel:
					if _, err := p.handleMessage(context.Background(), msg); err != nil {
						if _, ok := err.(*store.ErrInternal); ok {
							select {
							case p.errChan <- err:
							case <-p.stopping:
							}
							return
						}
					}
				}
			}
		}()
	}
}

func (p *Pipelined) handleMessage(ctx context.Context, msg interface{}) (hadPipelines bool, fErr error) {
	begin := time.Now()
	defer func() {
		duration := time.Since(begin)
		status := metricspkg.StatusLabelSuccess
		if fErr != nil {
			status = metricspkg.StatusLabelError
		}
		hasPipelines := "0"
		if hadPipelines {
			hasPipelines = "1"
		}
		messageHandlerDuration.
			WithLabelValues(status, hasPipelines).
			Observe(float64(duration) / float64(time.Millisecond))
	}()

	getter, ok := msg.(PipelineGetter)
	if !ok {
		panic("message received was not a PipelineGetter")
	}

	fields := getter.LogFields(false)
	pipelineRefs := getter.GetPipelines()

	// Add a legacy pipeline "reference" if msg is a
	// corev2.Event & has handlers.
	if event, ok := msg.(*corev2.Event); ok {
		if event.HasHandlers() {
			pipelineRefs = append(pipelineRefs, pipeline.LegacyPipelineReference())
		} else {
			logger.WithFields(fields).Debug("event has no handlers defined, skipping addition of legacy pipeline reference")
		}
	}

	if len(pipelineRefs) == 0 {
		logger.WithFields(fields).Info("no pipelines defined in resource")
		return false, nil
	}

	// loop through list of pipeline references and find
	// adapters that can run each of them.
	for _, ref := range pipelineRefs {
		adapterFound := false

		fields["pipeline_reference"] = ref.ResourceID()

		for _, adapter := range p.adapters {
			fields["pipeline_adapter"] = adapter.Name()

			if adapter.CanRun(ref) {
				if err := adapter.Run(ctx, ref, msg); err != nil {
					if _, ok := err.(*store.ErrInternal); ok {
						return true, err
					}
					skipPipelineErr := fmt.Errorf("%w, skipping execution of pipeline", err)
					if _, ok := err.(*pipeline.ErrNoWorkflows); ok {
						if fields["check_name"] != "keepalive" {
							// only warn about empty pipelines if it's not a keepalive.
							// this reduces log spam.
							logger.WithFields(fields).Warn(skipPipelineErr)
						}
					} else {
						logger.WithFields(fields).Error(skipPipelineErr)
					}
				}
				adapterFound = true
			}
		}
		if !adapterFound {
			return true, fmt.Errorf("no pipeline adapters were found that support the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
		}
	}

	return true, nil
}
